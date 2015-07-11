package main

import(
	"encoding/xml"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"

	"net/url"
	"log"
	"io/ioutil"
	"html/template"
	"strconv"
)

const (
	SizeSmallSquare = "s"
	SizeThumbnail   = "t"
	SizeSmall       = "m"
	SizeMedium500   = "-"
	SizeMedium640   = "z"
	SizeLarge       = "b"
	SizeOriginal    = "o"
	DatabaseName    = "puppies.sqlite"
)

const (
	FlickrEndPoint = "https://api.flickr.com/services/rest"
	FlickrQuery    = "flickr.photos.search"
	FlickrKey      = "60c30ef17e8d7721d395b8f158b1709f"
	PathPrefix     = "/pups"
	TopPupsPrefix  = "/top"
)

// Response for photo search requests.
type SearchResponse struct {
	Page    string  `xml:"page,attr"`
	Pages   string  `xml:"pages,attr"`
	PerPage string  `xml:"perpage,attr"`
	Total   string  `xml:"total,attr"`
	Photos  []Photo `xml:"photo"`
}

// Represents a Flickr photo.
type Photo struct {
	ID          string `xml:"id,attr"`
	Owner       string `xml:"owner,attr"`
	Secret      string `xml:"secret,attr"`
	Server      string `xml:"server,attr"`
	Farm        string `xml:"farm,attr"`
	Title       string `xml:"title,attr"`
	IsPublic    string `xml:"ispublic,attr"`
	IsFriend    string `xml:"isfriend,attr"`
	IsFamily    string `xml:"isfamily,attr"`
	Thumbnail_T string `xml:"thumbnail_t,attr"`
	Large_T     string `xml:"large_t,attr"`
}

type Vote struct {
	ID string `json:"id"`
	VT bool   `json:"vt"`
}
// Returns the URL to this photo in the specified size.
func (p *Photo) URL(size string) string {
	if size == "-" {
		return fmt.Sprintf("http://farm%s.static.flickr.com/%s/%s_%s.jpg",
			p.Farm, p.Server, p.ID, p.Secret)
	}
	return fmt.Sprintf("http://farm%s.static.flickr.com/%s/%s_%s_%s.jpg",
		p.Farm, p.Server, p.ID, p.Secret, size)
}

type flickrError struct {
	Code string `xml:"code,attr"`
	Msg  string `xml:"msg,attr"`
}

type Image struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	Large     string `json:"large"`
	UpVotes   int    `json:"upvotes"`
	DownVotes int    `json:"downvotes"`
}

type PuppiesResponse struct {
	Page    int      `json:"page"`
	Pages   int      `json:"pages"`
	PerPage int      `json:"perpage"`
	Total   int      `json:"total"`
	Images  []*Image `json:"images"`
}

func ListPuppies(w http.ResponseWriter, r *http.Request) {
	page := mux.Vars(r)["page"]
	if page == "" {
		page = "1"
	}
	tags := "puppy,dogs,dog,cute,pugs"

	baseUrl, err := url.Parse(FlickrEndPoint)

	params := url.Values{}
	params.Add("method", FlickrQuery)
	params.Add("api_key", FlickrKey)
	params.Add("tags", tags)
	params.Add("per_page", "20")
	params.Add("page", page)
	params.Add("safe_search", "2")
	params.Add("group_id","603018@N22")
	params.Add("sort", "date-posted-asc")

	baseUrl.RawQuery = params.Encode()

	resp, err := http.Get(baseUrl.String())
	if err != nil {
		// handle error, send proper error response
		log.Println(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error, send proper error response
		log.Println(err)
	}

	flickrResponse := struct {
		Stat   string         `xml:"stat,attr"`
		Err    flickrError    `xml:"err"`
		Photos SearchResponse `xml:"photos"`
	}{}

	xml.Unmarshal([]byte(body), &flickrResponse)

	//stat := flickrResult.Stat
	//if stat is "ok"
	if flickrResponse.Stat != "ok" {
		println(flickrResponse.Err.Msg)
		//return error message
	}

	searchResponse := flickrResponse.Photos
	flickrPhotos := searchResponse.Photos

	var tempIDs []string
	imageManager := NewImageManager()
	for _, ph := range flickrPhotos {
		img := imageManager.NewImage(ph)
		imageManager.Save(img)
		tempIDs = append(tempIDs, img.ID)
	}

	dbError := imageManager.InitDB(false)
	if dbError != nil {
		log.Printf("%q\n", dbError)
		return
	}
	defer imageManager.GetDB().Close()

	all := imageManager.All()

	dbPuppies := imageManager.FindOldPuppies(tempIDs)

	var newPuppies []*Image

	if len(dbPuppies) == 0 {
		imageManager.InsertPuppies(all)
	} else {
		for _, puppy := range dbPuppies {
			id := puppy.ID
			for _, allP := range all {
				//allPID, _ := strconv.Atoi(allP.ID)
				if allP.ID == id {
					allP.DownVotes = puppy.DownVotes
					allP.UpVotes = puppy.UpVotes
				} else {
					exists := true
					var existingPuppy *Image
					for _, np := range newPuppies {
						//nPID, _ := strconv.Atoi(np.ID)
						if np.ID == allP.ID {
							exists = false
							existingPuppy = allP
							break
						}
					}
					if exists == false {
						newPuppies = append(newPuppies, existingPuppy)
					}
				}
			}
		}
		imageManager.InsertPuppies(newPuppies)
	}


	puppiesResponse := imageManager.GetPuppiesResponse(&searchResponse)
	//response, err := json.Marshal(puppiesResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//w.Header().Set("Content-Type", "application/json")
	//w.Write(response)
	templates.ExecuteTemplate(w,"pups.html",puppiesResponse)
}

func UpdatePuppy(w http.ResponseWriter, r *http.Request) {
	var v Vote
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		println("some error")
	}

	imageManager := NewImageManager()

	dbError := imageManager.InitDB(false)
	if dbError != nil {
		log.Printf("%q\n", dbError)
		return
	}

	defer imageManager.GetDB().Close()
	id, err := strconv.Atoi(v.ID)
	imageManager.UpdateVotes(id, v.VT)

	response, err := json.Marshal(v)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func cloneImage(i *Image) *Image {
	c := *i
	return &c
}

//template concatenator kind of
var templates = template.Must(template.ParseFiles("edit.html", "pups.html"))


func main(){
	imageManager := NewImageManager()
	dbError := imageManager.InitDB(false)

	defer imageManager.GetDB().Close()

	if dbError != nil {
		log.Printf("%q\n", dbError)
		return
	} else {
		imageManager.CreateTables()
	}

	r := mux.NewRouter().StrictSlash(false)
	
	pups := r.Path(PathPrefix).Subrouter()
	pups.Methods("GET").HandlerFunc(ListPuppies)

	//update puppy likes/dislikes
	pupsUpdate := r.Path(PathPrefix).Subrouter()
	pupsUpdate.Methods("PUT").HandlerFunc(UpdatePuppy)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../../../content/")))
	http.Handle("/", r)
	http.ListenAndServe(":8080",nil)
}