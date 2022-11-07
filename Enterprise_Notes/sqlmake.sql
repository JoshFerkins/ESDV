DROP TABLE IF EXISTS User_Shares_T, Note_User_T, Note_T, User_T;

CREATE TABLE User_T (
    userID bigint NOT NULL,
    userName varchar(50) NOT NULL UNIQUE,
    userPass varchar(15) NOT NULL,
    userAuth smallint NOT NULL DEFAULT 1, 
    userPhone bigint DEFAULT 0,
    PRIMARY KEY (userID)
);

CREATE TABLE Note_T (
    noteID bigint NOT NULL,
    noteTitle varchar(200) NOT NULL,
    noteText TEXT DEFAULT ' ',
    createDateTime varchar(20) NOT NULL,
    compDateTime varchar(20),
    statusFlag int NOT NULL DEFAULT 5,
    ownedUser bigint NOT NULL,
    PRIMARY KEY (noteID),
    FOREIGN KEY (ownedUser) REFERENCES User_T(userID) 
);

CREATE TABLE Note_User_T (
    noteID bigint NOT NULL,
    userID bigint NOT NULL,
    permLevel int NOT NULL DEFAULT 1,
    FOREIGN KEY (noteID) REFERENCES Note_T(noteID) ON DELETE CASCADE,
    FOREIGN KEY (userID) REFERENCES User_T(userID) ON DELETE CASCADE
);

CREATE TABLE User_Shares_T(
    shareName VARCHAR(30),
    mainID BIGINT NOT NULL REFERENCES User_T(userID),
    friendID BIGINT NOT NULL REFERENCES User_T(userID)
);

INSERT INTO User_T (userID, userName, userPass, userAuth, userPhone)
VALUES 
        (543839, 'Josh Antony Ferkins', 'Ferkins123',  4, 02102884164),
        (193984, 'Tony Entity', 'gustav', 1, 0272800462),
        (12985437, 'WardenViel', 'warden', 2, 0214576677);

INSERT INTO Note_T(noteID, noteTitle, noteText, createDateTime, statusFlag, ownedUser)
VALUES 
        (123456, 'Testing Note 1', 'lorem ipsum 1', '01/02 5:24am', 1, 'Josh Antony Ferkins'),
        (584329, 'Testing Note 2', 'lorem ipsum 2', '02/03 6:35am', 2, 'Tony Entity'),
        (589321987, 'Testing note 3', 'lorem ipsum 3', '03/04 7:46am', 1, 'WardenViel');

INSERT INTO Note_User_T(noteID, userID, permLevel)
VALUES 
        (123456, 543839, 4),
        (123456, 193984, 1);

