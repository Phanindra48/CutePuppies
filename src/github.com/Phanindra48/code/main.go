package main

import(
	"encoding/xml"
	//"encoding/json"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"

	"net/url"
	"strconv"
	"log"
	"io/ioutil"
	"html/template"
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


type ImageManager struct {
	images []*Image
}

func NewImageManager() *ImageManager {
	return &ImageManager{}
}

func (m *ImageManager) GetPuppiesResponse(searchResponse *SearchResponse) *PuppiesResponse {
	page, err := strconv.Atoi(searchResponse.Page)
	pages, err := strconv.Atoi(searchResponse.Pages)
	perPage, err := strconv.Atoi(searchResponse.PerPage)
	total, err := strconv.Atoi(searchResponse.Total)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &PuppiesResponse{page, pages, perPage, total, m.images}
}

func (m *ImageManager) Save(image *Image) error {
	for _, im := range m.images {
		if im.ID == image.ID {
			return nil
		}
	}

	m.images = append(m.images, cloneImage(image))
	return nil
}

func cloneImage(i *Image) *Image {
	c := *i
	return &c
}

func (m *ImageManager) NewImage(photo Photo) *Image {
	return &Image{photo.ID, photo.Title, photo.URL(SizeThumbnail), photo.URL(SizeOriginal), 0, 0}
}

func ListPuppies(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w,"Puppy")
	tags := "puppy,dogs,dog,cute"
	page := "1"
	baseUrl, err := url.Parse(FlickrEndPoint)

	params := url.Values{}
	params.Add("method", FlickrQuery)
	params.Add("api_key", FlickrKey)
	params.Add("tags", tags)
	params.Add("per_page", "20")
	params.Add("page", page)
	params.Add("safe_search", "2")
	params.Add("group_id","70557968@N00")
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

	puppiesResponse := imageManager.GetPuppiesResponse(&searchResponse)
	//response, err := json.Marshal(puppiesResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//w.Header().Set("Content-Type", "application/json")
	//w.Write(response)
	templates.ExecuteTemplate(w,"pups.html",puppiesResponse)
}



//template concatenator kind of
var templates = template.Must(template.ParseFiles("edit.html", "pups.html"))


func main(){
	//imageManager := NewImageManager()
	r := mux.NewRouter().StrictSlash(false)
	
	pups := r.Path(PathPrefix).Subrouter()
	pups.Methods("GET").HandlerFunc(ListPuppies)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../../../content/")))
	http.Handle("/", r)
	http.ListenAndServe(":8080",nil)
}