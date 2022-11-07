package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

//----------------------------------------------------------------

//Managing login and registration

func loadLoginHandle(w http.ResponseWriter, r *http.Request) {
	//Retrieve parameters
	params := mux.Vars(r)
	var logindata LoginData

	//convert "v" to an integer
	val, err := strconv.Atoi(params["v"])
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	//retrieve array of user information
	userArr = userGetArr()

	//init struct with param variables
	logindata.UserName = params["uname"]
	logindata.Password = params["upass"]
	logindata.Confirmation = ""

	//If login event
	if len(logindata.UserName) > 0 && val == 1 {
		if logindata.Password == "null" {
			logindata.Confirmation = "Password must be entered"
			login_v.Render(w, logindata)
			return
		}
		if len(userArr) > 0 && val == 1 {
			for _, tempuser := range userArr {
				if tempuser.userName == logindata.UserName {
					if tempuser.userPass == logindata.Password {
						logindata.Confirmation = strconv.FormatUint(tempuser.userID, 10)
						currentUser = tempuser
						viewdata = noteGetArr()
						http.Redirect(w, r, "/index/", http.StatusFound)
						return
					} else {
						logindata.Confirmation = "Confirmation: Password incorrect"
					}
				}
			}
			if logindata.Confirmation == "" {
				logindata.Confirmation = "Confirmation: Username not found"
			}
		}
	} else if len(logindata.UserName) > 0 && val == 2 {
		//If register event
		if logindata.Password == "null" {
			logindata.Confirmation = "Password must be entered"
			login_v.Render(w, logindata)
			return
		}
		status := createUser(logindata.UserName, logindata.Password)
		if status {
			http.Redirect(w, r, "/index/", http.StatusFound)
			return
		} else {
			logindata.Confirmation = "Username must be: Unique, four or more characters, and musn't include numbers"
			login_v.Render(w, logindata)
			return
		}
	}
	login_v.Render(w, nil)
}

//----------------------------------------------------------------

//Loading index data, including boolean filter (status)
func loadIndexHandle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	filter := params["filter"]
	if len(filter) > 0 {
		viewdata.Notes = noteFilterArr(filter, viewdata.Notes)
	}
	index_v.Render(w, viewdata)
	viewdata = noteGetArr()
}

//Load a custom input filter (title, content, owner)
func filterIndexHandle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	r.ParseForm()

	v := params["filter"]
	if len(v) > 0 {
		viewdata.Notes = noteFilterArrCustom(v, viewdata.Notes, r.FormValue(v))
	}

	http.Redirect(w, r, "/index/", http.StatusFound)
}

//----------------------------------------------------------------

//Display note details
func editNoteHandle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	v, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	note := noteGet(v, viewdata)

	if note.StatusFlag == "Completed" || note.StatusFlag == "Cancelled" || (note.Owner != currentUser.userName && !checkUser(note)) {
		http.Redirect(w, r, "/index/", http.StatusFound)
	}

	editnote_v.Render(w, note)
}

//Accept and submit input for editing the note
func submitEditHandle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	r.ParseForm()

	v, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	details := EditDetails{v, r.FormValue("editTitle"), r.FormValue("editContent"), r.FormValue("statusSelect")}
	note := noteGet(v, viewdata)

	//Only a note owner may edit user settings
	if note.checkOwner() {
		note.noteAddUsers(r.FormValue("addUsers"), r.FormValue("permSelect"))
		note.noteRemoveUser(r.FormValue("removeUsers"))
	}

	if len(r.FormValue("saveUsers")) > 0 {
		note.saveUserShares(r.FormValue("saveUsers"))
	}

	editNote(details)
	http.Redirect(w, r, "/index/", http.StatusFound)
}

//----------------------------------------------------------------

//Create new note
func createNoteHandle(w http.ResponseWriter, r *http.Request) {
	createnote_v.Render(w, nil)
}

//Submitted info via html forms
func submitNoteHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	title := r.FormValue("createTitle")
	content := r.FormValue("createContent")
	createNote(title, content)
	viewdata = noteGetArr()
	http.Redirect(w, r, "/index/", http.StatusFound)
}

//----------------------------------------------------------------

//Edit exsiting profile

//loadProfile() and changeProfile() have the same header, however each are a different method
func loadProfileHandle(w http.ResponseWriter, r *http.Request) {
	//Initial load
	profile_v.Render(w, nil)
}

//Submit editing details about changing profile settings(username, passsword, phone)
func changeProfileHandle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	r.ParseForm() //Allows for access to form variables

	//Specific field names are parsed via headers
	v := params["name"]

	switch v {
	case "newUsername": //Changing username
		n := r.FormValue(v)
		currentUser.userName = n
		editUser(v, n)
	case "newPassword": //Changing password
		p := r.FormValue(v)
		currentUser.userPass = p
		editUser(v, p)
	case "newPhonenum": //Changing phone bumber
		pn := r.FormValue(v)
		//Requires indirect init of current user attribute, didn't like it since we had to create 'err' next to it
		pni, err := strconv.Atoi(pn)
		if err != nil {
			log.Fatal(err)
		}
		currentUser.userPhone = pni
		editUser(v, pn)
	}

	//Currently best way to solve disppearing html issue
	http.Redirect(w, r, "/profile/", http.StatusFound)
}
