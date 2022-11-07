package main

import "database/sql"

//HOnestly dunno which User I am using currently so...
type User struct {
	//userID is unique and created via a factory class
	userID   uint64
	userName string
	userPass string
	//userAuth is set to 1 upon initial creation
	/*
		userAuth
		1 == basic permissions
		2 == managerial permissions
		3 == administrative permissions
		4 == executive permissions
		none == unrecognised
	*/
	userAuth   int
	userPhone  int
	userShares []SharedPreset
}

type UserJSON struct {
	UserID     uint64 `json:"userID"`
	UserName   string `json:"userName"`
	UserAuth   string `json:"userAuth"`
	UserPhone  string `json:"userPhone"`
	UserPerm   string `json:"userPerm"`
	UserShares []SharedPreset
}

//Not being used
type UserPerm struct {
	noteID    uint64
	userID    uint64
	permLevel int
	/*
		1. view
		2. edit
		3. exec
	*/
}

//Depicts the base not structure to be accessed by the user
//Not using this one, using NoteJSON
type Note struct {
	//noteID is unique and created via a factory class
	noteID uint64

	//Strings made this format for readability purposes
	title          string
	text           string
	createDateTime string
	compDateTime   string
	owner          string

	statusFlag string
	/*
		Flag is to determine the current state of the note

		statusFlag
		1 == In Progress
		2 == completecmd
		3 == cancelled
		5 == halted
		none == unrecognised
	*/
}

//Using This one not regular NOTE
type NoteJSON struct {
	NoteID         uint64         `json:"noteID"`
	Title          string         `json:"title"`
	Text           sql.NullString `json:"text"`
	CreateDateTime string         `json:"createDateTime"`
	CompDateTime   sql.NullString `json:"compDateTime"`
	Owner          string         `json:"owner"`
	StatusFlag     string         `json:"statusFlag"`
	Users          []UserJSON
}

//For holding and displaying note data, could add a user array idk
type ViewData struct {
	Notes []NoteJSON
}

//Holds saved user share presets for each user
type SharedPreset struct {
	PresetName string `json:"presetName"`
	FriendID   uint64 `json:"friendID"`
}

//Holds temporary editing details for notes
type EditDetails struct {
	noteID uint64
	title  string
	text   string
	status string
}

//Holds temporary data for managing login
type LoginData struct {
	UserName     string
	Password     string
	Confirmation string
}
