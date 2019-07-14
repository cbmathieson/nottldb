DROP TABLE "favorites";
DROP TABLE "notes";
DROP TABLE "users";

CREATE TABLE "users" (
    pid BIGSERIAL PRIMARY KEY NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    --pass 
    profileImg VARCHAR(1024) NOT NULL,
    username VARCHAR(20) NOT NULL UNIQUE
);

--not referencing map table for lat/lon since it will deleted after a period of time
--dateFound will tell us when to delete from map
--myNotes on user's profile will be found using author_id 

--***Reference to user table so we can grab the up to date profile image + username***
--***Also can find who favorited it by taking instances of note_id in the favorites table***
-- HEADS UP, DATE CREATED TIMESTAMP IS NOT IN THE FORM WE'LL NEED INSIDE THE APP
CREATE TABLE "notes" (
    pid BIGSERIAL PRIMARY KEY NOT NULL,
    caption VARCHAR(30),
    dateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    dateFound TIMESTAMP,
    isAnonymous BOOLEAN NOT NULL,
    latitude VARCHAR(20) NOT NULL,
    longitude VARCHAR(20) NOT NULL,
    noteImage VARCHAR(1024) NOT NULL,
    author_id BIGINT NOT NULL,
    FOREIGN KEY(author_id) REFERENCES users(pid) ON DELETE CASCADE
);

--can count favorites from this table for each note
--can also build the list of favorites on user profile page
CREATE TABLE "favorites" (
    user_id BIGINT NOT NULL,
    note_id BIGINT NOT NULL,
    PRIMARY KEY (user_id,note_id),
    FOREIGN KEY (user_id) REFERENCES users(pid) ON DELETE CASCADE,
    FOREIGN KEY (note_id) REFERENCES notes(pid) ON DELETE CASCADE
);