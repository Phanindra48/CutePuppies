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
	//"github.com/jarias/stormpath-sdk-go"
	//"strings"
	"os"
	"github.com/gorilla/securecookie"
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
	UpdatePups	   = "/updatevotes"
	TopPupsPrefix  = "/top"
	viewsPath	   = "../../../content/views/"
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
	ID int   `json:"id"`
	VT int   `json:"vt"`
	UID int  `json:"uid"`
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
	UserChoice int   `json:"userchoice"`
}

type PuppiesResponse struct {
	Page    int      `json:"page"`
	Pages   int      `json:"pages"`
	PerPage int      `json:"perpage"`
	Total   int      `json:"total"`
	Images  []*Image `json:"images"`
}

type params struct {
	Page int `json:"page"`
	Uid int `json:"uid"`
}

func ListPuppies(w http.ResponseWriter, r *http.Request) {
	var t params
    contents, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Printf("%s", err)
        os.Exit(1)
    }
    fmt.Printf("%s\n", string(contents))
    jsonStr := string(contents)
    json.Unmarshal([]byte(jsonStr),&t)

    fmt.Println(t.Page,t.Uid)
	if t.Page == 0 {
		t.Page = 1
	}
	tags := "puppy,dogs,dog,cute,pugs"

	baseUrl, err := url.Parse(FlickrEndPoint)
	page := strconv.Itoa(t.Page)
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

	dbPuppies := imageManager.FindOldPuppies(tempIDs,t.Uid)

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
	w.Header().Set("Content-Type", "text/html")
	//w.Write(response)
	templates.ExecuteTemplate(w,"pups.html",puppiesResponse)
}

func ListTopPuppies(w http.ResponseWriter, r *http.Request) {
	fmt.Println("top puppies")
	var t params
    contents, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Printf("%s", err)
        os.Exit(1)
    }
    fmt.Printf("contents - > %s\n", string(contents))
    jsonStr := string(contents)
    json.Unmarshal([]byte(jsonStr),&t)

	imageManager := NewImageManager()
	dbError := imageManager.InitDB(false)
	if dbError != nil {
		log.Printf("%q\n", dbError)
		return
	}

	defer imageManager.GetDB().Close()
	//fmt.Printf("page-> %s\n",t.page)
	if t.Page == 0 {
		t.Page = 1
	}
	/*
	pageInt, err := strconv.Atoi(t.page)
	if err != nil {
		pageInt = 0
	}
	*/
	//pageInt := t.page


	puppies := imageManager.GetPuppiesByMostVotes(t.Page,t.Uid)
	count := imageManager.GetPuppiesCount()

	perPage := 20
	pages := count / perPage

	searchResponse := PuppiesResponse{t.Page, pages, perPage, count, puppies}

	//response, err := json.Marshal(searchResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w,"pups.html",searchResponse)
}

func UpdatePuppy(w http.ResponseWriter, r *http.Request) {
	var v Vote
    contents, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Printf("%s", err)
        os.Exit(1)
    }
    fmt.Printf("%s\n", string(contents))
    jsonStr := string(contents)
    json.Unmarshal([]byte(jsonStr),&v)
    //fmt.Printf("update puppy photo id %d,vote %v\n",v.ID,v.VT)

	imageManager := NewImageManager()

	dbError := imageManager.InitDB(false)
	if dbError != nil {
		log.Printf("dbError -> %q\n", dbError)
		return
	}

	defer imageManager.GetDB().Close()
	//id, err := strconv.Atoi(v.ID)
	choice := v.VT == 1
	fmt.Printf("update puppy photo id %d,vote %v,uid: %d\n",v.ID,v.VT,v.UID)
	imageManager.UpdateVotes(v.ID, choice,v.UID)

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
var templates = template.Must(
	template.ParseFiles(viewsPath + "pups.html",viewsPath + "login.html",viewsPath + "home.html"))


//login users
var cookieHandler = securecookie.New(
    securecookie.GenerateRandomKey(64),
    securecookie.GenerateRandomKey(32))
 
func indexPageHandler(response http.ResponseWriter, request *http.Request) {
	userName,userId := getUserName(request)
	//userName := "werew"
	if userName != "" && userId > 0 {
	    //fmt.Fprintf(response, internalPage, userName)
	    templates.ExecuteTemplate(response,"home.html",userId)
	} else {
	    templates.ExecuteTemplate(response,"login.html","")
	}
}
/*
func signupHandler(response http.ResponseWriter, request *http.Request){
	email := request.FormValue("email")
	pass := request.FormValue("password")
	username := request.FormValue("username")

}
*/
func loginHandler(response http.ResponseWriter, request *http.Request) {
	name := request.FormValue("email")
	pass := request.FormValue("password")
	//fmt.Printf("Logged in %s %s \n",name,pass);
	redirectTarget := "/"
	if name != "" && pass != "" {
	    // .. check credentials ..
	    setSession(name, response)
	    redirectTarget = "/"
	    //fmt.Printf("Logged in %s \n",name);
	}
	http.Redirect(response, request, redirectTarget, 302)
}
 
func logoutHandler(response http.ResponseWriter, request *http.Request) {
 	fmt.Printf("logout \n")
    clearSession(response)
    http.Redirect(response, request, "/", 302)
 }

func setSession(userName string, response http.ResponseWriter) {
    value := map[string]string{
         "name": userName,
    }
    if encoded, err := cookieHandler.Encode("session", value); err == nil {
        cookie := &http.Cookie{
            Name:  "session",
            Value: encoded,
            Path:  "/",
        }
        http.SetCookie(response, cookie)
        insertUser(userName)
        fmt.Println("set session");
    } else{
     	fmt.Println("set session panic")
    }
 }
 
func getUserName(request *http.Request) (userName string,user_id int) {
    if cookie, err := request.Cookie("session"); err == nil {
        cookieValue := make(map[string]string)
        if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
            userName = cookieValue["name"]
            //fmt.Println("cookie handler")
        }
    }
    fmt.Println("getUserName")
    imageManager := NewImageManager()
    dbError := imageManager.InitDB(false)
	if dbError != nil {
		log.Fatal(dbError)

	}
	defer imageManager.GetDB().Close()
    if len(userName) > 0 {
    	user_id = imageManager.getUser(userName)
    	fmt.Printf("getUsername method - user id %d\n",user_id)
	}
    
    return userName,user_id
}

func clearSession(response http.ResponseWriter) {
     cookie := &http.Cookie{
         Name:   "session",
         Value:  "",
         Path:   "/",
         MaxAge: -1,
     }
     http.SetCookie(response, cookie)
}


func insertUser(username string) {
	imageManager := NewImageManager()

	dbError := imageManager.InitDB(false)
	if dbError != nil {
		log.Printf("dbError -> %q\n", dbError)
		return
	}

	defer imageManager.GetDB().Close()
	imageManager.InsertUser(username)
}

func main(){
	
	r := mux.NewRouter().StrictSlash(false)
	
	r.HandleFunc("/", indexPageHandler)
	r.HandleFunc("/login",loginHandler).Methods("POST")
	r.HandleFunc("/logout",logoutHandler).Methods("POST")

	imageManager := NewImageManager()
	dbError := imageManager.InitDB(false)

	defer imageManager.GetDB().Close()

	if dbError != nil {
		log.Printf("dbError -> %q\n", dbError)
		return
	} else {
		imageManager.CreateTables()
	}
	pups := r.Path(PathPrefix).Subrouter()
	pups.Methods("POST").HandlerFunc(ListPuppies)

	topPups := r.Path(TopPupsPrefix).Subrouter()
	topPups.Methods("POST").HandlerFunc(ListTopPuppies)

	//update puppy likes/dislikes
	pupsUpdate := r.Path(UpdatePups).Subrouter()
	pupsUpdate.Methods("POST").HandlerFunc(UpdatePuppy)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../../../content/")))
	http.Handle("/", r)
	http.ListenAndServe(":8080",nil)
}