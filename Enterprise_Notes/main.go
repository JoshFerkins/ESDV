package main

import (
	"context"
	"database/sql"
	_ "errors"
	"fmt" //pronounced f-mmm-t
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
	"unicode"

	//for use with postgres database
	//"database/sql"

	/*sony is used for creating unique identifiers for
	mutiple aspects of the program; User, and Note*/

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sony/sonyflake"
)

//Global Variables

//Database variables
const (
	host     = "localhost" //"server ip address for portable postgresql"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "ESDVdb"
)

//General postgres db info
var psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

//used for id generation
var flake *sonyflake.Sonyflake

//assign currentUser
var currentUser User

//----------------------------------------------------------------------------------

/*
				[][]              [][]
				||||              ||||
				||||              ||||
				||||              ||||
				||||              ||||
				||||              ||||
				||||              /|||
				\\\\             ////
				 \\\\___________////
  				\\[===========]//
*/

//USER SECTION

//----------------------------------------------------------------------------------
//Standard array to hold all users
var userArr []User

//Return an entire array of every user
func userGetArr() (arr []User) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement := "SELECT * FROM User_T"
	rows, err := db.Query(statement)
	if err != nil {
		log.Fatal(err)
	}

	var tempArr []User
	var tempUser User

	for rows.Next() {
		err = rows.Scan(&tempUser.userID, &tempUser.userName, &tempUser.userPass, &tempUser.userAuth, &tempUser.userPhone)
		if err != nil {
			return
		}
		tempUser.userShares = userGetShares(tempUser.userID)
		tempArr = append(tempArr, tempUser)
	}
	return tempArr
}

//Return all user share presents a user has made
func userGetShares(id uint64) (u []SharedPreset) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement := "SELECT shareName, friendID FROM User_Shares_T WHERE mainID = " + strconv.FormatUint(id, 10)

	rows, err := db.Query(statement)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var name string
		var friendID uint64

		err = rows.Scan(&name, &friendID)
		if err != nil {
			log.Fatal(err)
		} else if err == nil {
			return
		}

		preset := SharedPreset{name, friendID}
		fmt.Println(preset)
		u = append(u, preset)
	}

	return u
}

//Convert auth, phone, and perm to strings
func convUserAttr(auth int, phone int, perm int) (string, string, string) {
	authConv := ""
	switch auth {
	case 1:
		authConv = "Basic"
	case 2:
		authConv = "Manager"
	case 3:
		authConv = "Admin"
	case 4:
		authConv = "Exec"
	}

	phoneConv := strconv.Itoa(phone)

	permConv := ""
	switch perm {
	case 1:
		permConv = "View"
	case 2:
		permConv = "Edit"
	case 3:
		permConv = "Owner"
	}

	return authConv, phoneConv, permConv
}

//Change user details
func editUser(v string, newvalue string) (status bool) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var statement string

	switch v {
	case "newUsername":
		if !userNameValidity(newvalue) {
			return false
		}
		statement = "UPDATE User_T SET userName = '" + newvalue + "' WHERE userName LIKE '" + currentUser.userName + "'"
	case "newPassword":
		statement = "UPDATE User_T SET userPass = '" + newvalue + "' WHERE userName LIKE '" + currentUser.userName + "'"
	case "newPhonenum":
		statement = "UPDATE User_T SET  userPhone = '" + newvalue + "' WHERE userName LIKE '" + currentUser.userName + "'"
	}
	_, err = db.Exec(statement)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

//Creates user via user input
func createUser(uname string, upass string) (status bool) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if !userNameValidity(uname) {
		return false
	}
	id := setID()
	//TestCreateUser(user)
	statement := "INSERT INTO User_T (userID, userName, userPass, userAuth, userPhone) VALUES ($1, $2, $3, $4, $5)"
	_, err = db.Exec(statement, id, uname, upass, 1, 0)
	if err != nil {
		log.Fatal(err)
	}

	statement = "SELECT * FROM User_T WHERE userName LIKE '" + uname + "'"

	row := db.QueryRow(statement)

	switch err = row.Scan(&currentUser.userID, &currentUser.userName, &currentUser.userPass, &currentUser.userAuth, &currentUser.userPhone); err {
	case sql.ErrNoRows:
		return false
	case nil:
		break
	default:
		fmt.Println("Sql error")
		panic(err)
	}

	return true
}

//Checks user validity
func userNameValidity(uname string) bool {

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement := "SELECT COUNT(*) FROM user_t WHERE userName LIKE '" + uname + "'"
	var count int

	err = db.QueryRow(statement).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count >= 1 {
		return false
	}

	if len(uname) < 4 {
		return false
	} else {
		for _, char := range uname {
			if unicode.IsNumber(char) {
				return false
			} else {
				return true
			}
		}
	}
	return true
}

//----------------------------------------------------------------------------

/*
			[[||\\       [][]
			||||\\\      ||||
      ||||\\\\     ||||
			|||| \\\\    ||||
			||||  \\\\   ||||
			||||   \\\\  ||||
			||||    \\\\ ||||
			||||     \\\\||||
			||||      \\\||||
			[][]       \\||]]
*/

//NOTE SECTION

//----------------------------------------------------------------------------

var viewdata ViewData

//Remove a user from a note
func (n NoteJSON) noteRemoveUser(name string) bool {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var userID uint64

	for _, user := range n.Users {
		if user.UserName == name {
			userID = user.UserID
		}
	}

	statement := "DELETE FROM Note_User_T WHERE userID = " + strconv.FormatUint(userID, 10)

	_, err = db.Exec(statement)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

//Adds users to specified note through the note edit page
func (n NoteJSON) noteAddUsers(name string, perm string) bool {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement := "SELECT COUNT(*) FROM User_T WHERE userName LIKE '" + name + "'"
	var count int

	err = db.QueryRow(statement).Scan(&count)

	if count != 1 {
		return false
	}

	if len(userArr) > 0 {
		for _, user := range userArr {
			if user.userName == name {
				statement = "INSERT INTO Note_User_T(noteID, userID, permLevel) VALUES($1, $2, $3)"
				_, err = db.Exec(statement, n.NoteID, user.userID, perm)
				if err != nil {
					log.Fatal(err)
				}
				return true
			}
		}
		return false
	} else {
		return false
	}
}

//Retrieves object of type ViewData with a filled array of notes and users with those notes
func noteGetArr() (v ViewData) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal()
	}
	defer db.Close()
	statement := "SELECT Note_T.* FROM Note_T INNER JOIN Note_User_T ON Note_User_T.noteID = Note_T.noteID INNER JOIN User_T ON User_T.userID = Note_User_T.userID AND User_T.userName LIKE '" + currentUser.userName + "'"
	rows, err := db.Query(statement)
	if err != nil {
		log.Fatal(err)
	}
	n := NoteJSON{}
	viewdata := ViewData{}
	for rows.Next() {
		var statusInt int
		var owner64 uint64

		err := rows.Scan(&n.NoteID, &n.Title, &n.Text, &n.CreateDateTime, &n.CompDateTime, &statusInt, &owner64)
		if err != nil {
			return
		}
		n.Users = noteGetUserArr(n.NoteID)
		n.Owner, n.StatusFlag = convOwnerStatus(owner64, statusInt, n.NoteID)

		viewdata.Notes = append(viewdata.Notes, n)
	}
	return viewdata
}

//Retrieves all users associated with this note
func noteGetUserArr(id uint64) []UserJSON {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal()
	}
	defer db.Close()

	statement := "SELECT User_T.userID, User_T.userName, User_T.userAuth, User_T.userPhone, Note_User_T.permLevel FROM User_T INNER JOIN Note_User_T ON Note_User_T.userID = User_T.userID INNER JOIN Note_T ON Note_T.noteID = Note_User_T.noteID	AND Note_T.noteID = " + strconv.FormatUint(id, 10)

	u := UserJSON{}
	var arr []UserJSON

	users, err := db.Query(statement)
	if err != nil {
		log.Fatal(err)
	}
	for users.Next() {
		var auth int
		var phone int
		var perm int

		err = users.Scan(&u.UserID, &u.UserName, &auth, &phone, &perm)
		if err != nil {
			log.Fatal(err)
		}

		u.UserAuth, u.UserPhone, u.UserPerm = convUserAttr(auth, phone, perm)

		arr = append(arr, u)
	}
	return arr
}

//global boolians
var statusAsc bool = true

func noteFilterArr(f string, arr []NoteJSON) []NoteJSON {

	switch f {

	//case "completed":

	case "status":
		statusMap := map[string]int{"In Progress": 1, "Completed": 2, "Cancelled": 3, "Halted": 4}

		//Sorting
		for i := 0; i < len(arr)-1; i++ {
			for j := 0; j < len(arr)-i-1; j++ {
				if statusAsc {
					if statusMap[arr[j].StatusFlag] > statusMap[arr[j+1].StatusFlag] {
						arr[j], arr[j+1] = arr[j+1], arr[j]
					}
				} else {
					if statusMap[arr[j].StatusFlag] < statusMap[arr[j+1].StatusFlag] {
						arr[j], arr[j+1] = arr[j+1], arr[j]
					}
				}
			}
		}

		//Change bool
		if statusAsc {
			statusAsc = false
		} else {
			statusAsc = true
		}
		return arr
	default:
		return arr
	}
}

//Filters the note array depending on custom inputted data (title, owner, content)
func noteFilterArrCustom(f string, arr []NoteJSON, c string) (newarr []NoteJSON) {
	for i := 0; i < len(arr); i++ {
		switch f {
		case "filterTitle":
			if strings.Contains(strings.ToLower(arr[i].Title), strings.ToLower(c)) {
				newarr = append(newarr, arr[i])
			}
		case "filterOwner":
			if arr[i].Owner == c {
				newarr = append(newarr, arr[i])
			}
		case "filterContent":
			if strings.Contains(strings.ToLower(arr[i].Text.String), strings.ToLower(c)) {
				newarr = append(newarr, arr[i])
			}
		case "reset":
			viewdata = noteGetArr()
			return viewdata.Notes
		default:
			newarr = arr
		}
	}
	return newarr
}

//Return a single note
func noteGet(id uint64, viewdata ViewData) (note NoteJSON) {
	for _, note := range viewdata.Notes {
		if note.NoteID == id {
			return note
		}
	}
	return
}

//Convert owner(uint64) and statusflag(int) to strings
func convOwnerStatus(owner64 uint64, statusInt int, noteID uint64) (string, string) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal()
	}
	defer db.Close()

	var owner string
	var status string

	statement := "SELECT userName FROM User_T JOIN Note_T ON User_T.userID = " + strconv.FormatUint(owner64, 10) + " AND noteID = " + strconv.FormatUint(noteID, 10)
	err = db.QueryRow(statement).Scan(&owner)
	if err != nil {
		log.Fatal(err)
	}

	statusMap := map[int]string{1: "In Progress", 2: "Completed", 3: "Cancelled", 4: "Halted"}
	status = statusMap[statusInt]

	return owner, status
}

//Create the main note
func createNote(title string, content string) bool {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal()
	}
	defer db.Close()

	if !titleValidation(title) {
		return false
	}

	noteID := setID()

	statement := "INSERT INTO Note_T(noteID, noteTitle, noteText, createDateTime, statusFlag, ownedUser) VALUES($1, $2, $3, $4, $5, $6)"

	_, err = db.Exec(statement, noteID, title, content, getDate(), 1, currentUser.userID)
	if err != nil {
		log.Fatal(err)
	}

	createAddOwner(noteID)

	return true
}

//Add the owner/creator as the owner in the database table
func createAddOwner(noteID uint64) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement := "INSERT INTO Note_User_T(noteID, userID, permLevel) VALUES($1, $2, $3)"

	_, err = db.Exec(statement, noteID, currentUser.userID, 3)
	if err != nil {
		log.Fatal(err)
	}
}

//Change note details
func editNote(details EditDetails) bool {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement := ""

	//Title validation
	if !titleValidation(details.title) {
		return false
	}

	//Status flag conversion

	//If the status is "complete"
	if details.status == "2" {
		//Adds a completed date & time
		statement = "UPDATE Note_T SET noteTitle = $1, noteText = $2, compDateTime = $3, statusFlag = $4 WHERE noteID = $5;"
		_, err = db.Exec(statement, details.title, details.text, getDate(), details.status, details.noteID)
		if err != nil {
			log.Fatal(err)
		}
	} else if details.status == "0" {
		statement = "UPDATE Note_T SET noteTitle = $1, noteText = $2 WHERE noteID = $5;"
		_, err = db.Exec(statement, details.title, details.text, details.noteID)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		statement = "UPDATE Note_T SET noteTitle = $1, noteText = $2, statusFlag = $3 WHERE noteID = $4;"
		_, err = db.Exec(statement, details.title, details.text, details.status, details.noteID)
		if err != nil {
			log.Fatal(err)
		}
	}

	return true
}

//Poor language usage coming up
func titleValidation(title string) bool {
	//Appologise for the poor english in this section
	badwords := []string{"fuck", "hell", "cunt", "shit", "bitch"} //Other bad words
	//Don't degrade me for this lmfao

	for i := 0; i < len(badwords); i++ {
		if strings.Contains(strings.ToLower(title), badwords[i]) {
			return false
		}
	}

	return true
}

//Checks if the current user is the owner of a note
func (n NoteJSON) checkOwner() bool {
	for _, user := range n.Users {
		if currentUser.userID == user.UserID && user.UserPerm == "Owner" {
			return true
		}
	}
	return false
}

//Check if user is appart of this note and has edit || owner permissions to edit
func checkUser(n NoteJSON) bool {
	for _, user := range n.Users {
		if user.UserID == currentUser.userID && (user.UserPerm == "Edit" || user.UserPerm == "Owner") {
			return true
		}
	}
	return false
}

//Save note users as a user share preset
func (n NoteJSON) saveUserShares(name string) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, user := range n.Users {
		if user.UserID != currentUser.userID {
			statement := "INSERT INTO User_Shares_T(shareName, mainID, friendID) VALUES($1, $2, $3)"
			_, err = db.Exec(statement, name, currentUser.userID, user.UserID)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

//----------------------------------------------------------------------------------
//Factory Settings

//Factory, this is designed to create and return a random uint64 id
func setID() uint64 {
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})

	id, err := flake.NextID()
	if err != nil {
		log.Fatal(err)
	}
	return id
}

//Factory function to return current date and time, (DDMMYY, HHMM)
func getDate() string {
	t := time.Now()
	return t.Format("2 Jan 06 03:04PM")
}

//----------------------------------------------------------------------------------

/*
	[[||\\               //||]]			[][]              [][]    \\\\        ////
	||||\\\             ///||||     ||||              ||||     \\\\      ////
	|||[\\\\           ////]|||			||||              ||||      \\\\    ////
	|||| \\\\         //// ||||			||||              ||||       \\\\  ////
	||||  \\\\       ////  ||||			||||              |||| 				\\\\////
	||||   \\\\     ////   ||||			||||              ||||			  ////\\\\
	||||    \\\\   ////    ||||			\\\\             ////        ////  \\\\
	||||     \\\\ ////     ||||      \\\\___________////        ////    \\\\
	[][]      \\|||//      [][]       \\{===========}//        ////      \\\\
*/

//MUX SERVER & ROUTING

//----------------------------------------------------------------------------------

//All code from down here was inspired and modified from the jobs.tradie program, this is not original content, but has been repurposed
const (
	BINDPORT string = "3000" //change to suit requirement
)

var wait time.Duration

var index_v *View
var login_v *View
var profile_v *View
var editnote_v *View
var createnote_v *View

//setup some route endpoints
func newRouter() *mux.Router {
	r := mux.NewRouter()

	//These subrouters aren't necessary, these are just for me to better manage routes, may affect efficiency idk
	loadRouter := r.PathPrefix("/").Subrouter()
	indexRouter := r.PathPrefix("/index").Subrouter()
	editRouter := r.PathPrefix("/edit").Subrouter()
	createRouter := r.PathPrefix("/create").Subrouter()
	profRouter := r.PathPrefix("/profile").Subrouter()

	//login handlers
	loadRouter.HandleFunc("/", loadLoginHandle).Methods("GET")
	loadRouter.HandleFunc("/{uname}/{upass}/{v:[0-9]+}", loadLoginHandle).Methods("GET")

	//Main notes index page
	indexRouter.HandleFunc("/", loadIndexHandle).Methods("GET")
	indexRouter.HandleFunc("/{filter}", loadIndexHandle).Methods("GET")    //Click filter (status)
	indexRouter.HandleFunc("/{filter}", filterIndexHandle).Methods("POST") //Input filter (title, owner, content)

	//Handle note functionality
	editRouter.HandleFunc("/{id}", editNoteHandle).Methods("GET") //Edit individual note
	editRouter.HandleFunc("/{id}", submitEditHandle).Methods("POST")
	createRouter.HandleFunc("/", createNoteHandle).Methods("GET") //Create new note
	createRouter.HandleFunc("/", submitNoteHandle).Methods("POST")

	//Profile handlers
	profRouter.HandleFunc("/", loadProfileHandle).Methods("GET")
	profRouter.HandleFunc("/{name}", changeProfileHandle).Methods("POST")
	//profRouter.HandleFunc("/modify", modifyProfile).Methods("GET")

	return r
}

//used to auto detect the active local IP address - not used yet
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func main() {

	//id, err := flake.NextID()
	//create the initial collection using some sample data, this only gets created
	//if the collection does not exist
	// load templates
	log.Println("Loading templates")
	index_v = NewView("bootstrap", "views/index.gohtml")
	login_v = NewView("bootstrap", "views/login.gohtml")
	profile_v = NewView("bootstrap", "views/profile.gohtml")
	editnote_v = NewView("bootstrap", "views/editnote.gohtml")
	createnote_v = NewView("bootstrap", "views/createnote.gohtml")

	log.Println("Starting HTTP service on " + BINDPORT)
	r := newRouter()

	// setup HTTP on gorilla mux for a gracefull shutdown
	srv := &http.Server{
		Addr: "0.0.0.0:" + BINDPORT,

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// HTTP listener is in a goroutine as its blocking
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// setup a ctrl-c trap to ensure a graceful shutdown
	// this would also allow shutting down other pipes/connections. eg DB
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
