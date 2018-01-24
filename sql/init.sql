/*
 * Plex Event Database
 * sqlite v3.21.0
*/

-- name: create-table-event
CREATE TABLE IF NOT EXISTS event(
  eId INTEGER PRIMARY KEY,
  date VARCHAR(25) NOT NULL -- RFC3339, f.e. 2006-01-02T15:04:05+07:00
    -- For performance I don't want to use RegEx in the constrain,
    -- so I decided to do a quick and simple check.
    CHECK (SUBSTR(date, 5, 1) IS '-' AND
           SUBSTR(date, 8, 1) IS '-' AND
           SUBSTR(date, 11, 1) IS 'T' AND
           SUBSTR(date, 14, 1) IS ':' AND
           SUBSTR(date, 17, 1) IS ':' AND
           SUBSTR(date, 20, 1) IN ('Z', '+', '-') AND
           SUBSTR(date, 23, 1) IS ':'), -- Note: Timezone information should always be included
  type VARCHAR(10) NOT NULL
    CHECK (type IN ('play', 'stop', 'pause', 'resume', 'userRating', 'scrobble')),
  rating UNSIGNED INTEGER(2)
    CHECK (rating <= 10),
  local BOOLEAN NOT NULL,
  owned BOOLEAN NOT NULL,

  accountNumber INTEGER(5) NOT NULL,
  sUUID VARCHAR(40) NOT NULL,
  cUUID VARCHAR(40) NOT NULL,
  mGUID VARCHAR(40) NOT NULL,
  aId NOT NULL,
  FOREIGN KEY(accountNumber) REFERENCES account,
  FOREIGN KEY(sUUID) REFERENCES server,
  FOREIGN KEY(cUUID) REFERENCES client,
  FOREIGN KEY(mGUID) REFERENCES media,
  FOREIGN KEY(aId) REFERENCES address
);

-- name: create-trigger-rating
CREATE TRIGGER IF NOT EXISTS ratingTrigger AFTER INSERT ON event -- When type isn't userRating we want rating to be NULL
  WHEN NEW.type IS NOT 'userRating' AND NEW.rating IS NOT NULL BEGIN
  UPDATE event SET rating = NULL WHERE eId = NEW.eId;
END;

-- name: create-table-media
CREATE TABLE IF NOT EXISTS media(
  guid VARCHAR(200) PRIMARY KEY NOT NULL,

  type VARCHAR(7) NOT NULL
    CHECK (type IN ('movie', 'episode', 'track', 'image', 'clip')),
  subtype VARCHAR(7),

  key VARCHAR(48),
  parentKey VARCHAR(48),
  grandparentKey VARCHAR(48),
  primaryExtraKey VARCHAR(48),

  title VARCHAR(255) NOT NULL,
  titleSort VARCHAR(255),
  parentTitle VARCHAR(255),
  grandparentTitle VARCHAR(255),

  summary TEXT,
  duration INTEGER,

  thumb VARCHAR(255),
  parentThumb VARCHAR(255),
  grandparentThumb VARCHAR(255),

  grandparentTheme VARCHAR(96),
  grandparentRatingKey INTEGER,

  art VARCHAR(255),
  grandparentArt VARCHAR(255),

  `index` INTEGER,
  parentIndex INTEGER,

  studio VARCHAR(64),
  tagline VARCHAR(255),
  chapterSource VARCHAR(10),

  librarySectionID INTEGER,
  librarySectionKey VARCHAR(40),
  librarySectionType VARCHAR(6),

  webRating REAL,
  userRating REAL,
  audienceRating REAL,
  contentRating VARCHAR(10),
  ratingImage VARCHAR(96),
  viewCount INTEGER,

  releaseYear INTEGER(4),
  dateOriginal VARCHAR(25) -- RFC3339, f.e. 2006-01-02T15:04:05+07:00
    -- For performance I don't want to use RegEx in the constrain,
    -- so I decided to do a quick and simple check.
    CHECK (dateOriginal IS '' OR (
      SUBSTR(dateOriginal, 5, 1) IS '-' AND
      SUBSTR(dateOriginal, 8, 1) IS '-' AND
      SUBSTR(dateOriginal, 11, 1) IS 'T' AND
      SUBSTR(dateOriginal, 14, 1) IS ':' AND
      SUBSTR(dateOriginal, 17, 1) IS ':' AND
      SUBSTR(dateOriginal, 20, 1) IN ('Z', '+', '-') AND
      (LENGTH(dateOriginal) IS 20 OR SUBSTR(dateOriginal, 23, 1) IS ':'))),
  -- Note: RFC3339 without timezone information end here
  dateAdded VARCHAR(25) -- RFC3339, f.e. 2006-01-02T15:04:05+07:00
    -- For performance I don't want to use RegEx in the constrain,
    -- so I decided to do a quick and simple check.
    CHECK (dateAdded IS '' OR (
      SUBSTR(dateAdded, 5, 1) IS '-' AND
      SUBSTR(dateAdded, 8, 1) IS '-' AND
      SUBSTR(dateAdded, 11, 1) IS 'T' AND
      SUBSTR(dateAdded, 14, 1) IS ':' AND
      SUBSTR(dateAdded, 17, 1) IS ':' AND
      SUBSTR(dateAdded, 20, 1) IN ('Z', '+', '-') AND
      (LENGTH(dateAdded) IS 20 OR SUBSTR(dateAdded, 23, 1) IS ':'))),
  -- Note: RFC3339 without timezone information end here
  dateUpdated VARCHAR(25) -- RFC3339, f.e. 2006-01-02T15:04:05+07:00
    -- For performance I don't want to use RegEx in the constrain,
    -- so I decided to do a quick and simple check.
    CHECK (dateUpdated IS '' OR (
      SUBSTR(dateUpdated, 5, 1) IS '-' AND
      SUBSTR(dateUpdated, 8, 1) IS '-' AND
      SUBSTR(dateUpdated, 11, 1) IS 'T' AND
      SUBSTR(dateUpdated, 14, 1) IS ':' AND
      SUBSTR(dateUpdated, 17, 1) IS ':' AND
      SUBSTR(dateUpdated, 20, 1) IN ('Z', '+', '-') AND
      (LENGTH(dateUpdated) IS 20 OR SUBSTR(dateUpdated, 23, 1) IS ':')))
  -- Note: RFC3339 without timezone information end here
);

-- name: create-table-filter
CREATE TABLE IF NOT EXISTS filter(
  fId INTEGER PRIMARY KEY NOT NULL,
  tag VARCHAR(32),
  filter VARCHAR(32),
  role VARCHAR(64),
  thumb VARCHAR(255),
  count INTEGER
);

-- name: create-table-server
CREATE TABLE IF NOT EXISTS server(
  uuid VARCHAR(40) PRIMARY KEY NOT NULL,
  name VARCHAR(50) NOT NULL
);

-- name: create-table-account
CREATE TABLE IF NOT EXISTS account(
  plexNumber INTEGER(5) PRIMARY KEY NOT NULL,
  name VARCHAR(100) NOT NULL,
  thumbnail VARCHAR(255)
);

-- name: create-table-client
CREATE TABLE IF NOT EXISTS client(
  uuid VARCHAR(40) PRIMARY KEY NOT NULL,
  name VARCHAR(255)
);

-- name: create-table-address
CREATE TABLE IF NOT EXISTS address(
  aId INTEGER PRIMARY KEY,
  ip VARCHAR(39) NOT NULL UNIQUE -- max(IPv6)
);

-- relations

-- name: create-table-hasDirector
CREATE TABLE IF NOT EXISTS hasDirector (
  guid VARCHAR(200) NOT NULL,
  fId INTEGER NOT NULL,
  FOREIGN KEY(guid) REFERENCES media,
  FOREIGN KEY(fId) REFERENCES filter,
  PRIMARY KEY(guid, fId)
);

-- name: create-table-hasProducer
CREATE TABLE IF NOT EXISTS hasProducer (
  guid VARCHAR(200) NOT NULL,
  fId INTEGER NOT NULL,
  FOREIGN KEY(guid) REFERENCES media,
  FOREIGN KEY(fId) REFERENCES filter,
  PRIMARY KEY(guid, fId)
);

-- name: create-table-isSimilarWith
CREATE TABLE IF NOT EXISTS isSimilarWith (
  guid VARCHAR(200) NOT NULL,
  fId INTEGER NOT NULL,
  FOREIGN KEY(guid) REFERENCES media,
  FOREIGN KEY(fId) REFERENCES filter,
  PRIMARY KEY(guid, fId)
);

-- name: create-table-hasWriter
CREATE TABLE IF NOT EXISTS hasWriter (
  guid VARCHAR(200) NOT NULL,
  fId INTEGER NOT NULL,
  FOREIGN KEY(guid) REFERENCES media,
  FOREIGN KEY(fId) REFERENCES filter,
  PRIMARY KEY(guid, fId)
);

-- name: create-table-hasRole
CREATE TABLE IF NOT EXISTS hasRole (
  guid VARCHAR(200) NOT NULL,
  fId INTEGER NOT NULL,
  FOREIGN KEY(guid) REFERENCES media,
  FOREIGN KEY(fId) REFERENCES filter,
  PRIMARY KEY(guid, fId)
);

-- name: create-table-hasGenre
CREATE TABLE IF NOT EXISTS hasGenre (
  guid VARCHAR(200) NOT NULL,
  fId INTEGER NOT NULL,
  FOREIGN KEY(guid) REFERENCES media,
  FOREIGN KEY(fId) REFERENCES filter,
  PRIMARY KEY(guid, fId)
);

-- name: create-table-isFromCountry
CREATE TABLE IF NOT EXISTS isFromCountry (
  guid VARCHAR(200) NOT NULL,
  fId INTEGER NOT NULL,
  FOREIGN KEY(guid) REFERENCES media,
  FOREIGN KEY(fId) REFERENCES filter,
  PRIMARY KEY(guid, fId)
);

-- name: create-table-isInCollection
CREATE TABLE IF NOT EXISTS isInCollection (
  guid VARCHAR(200) NOT NULL,
  fId INTEGER NOT NULL,
  FOREIGN KEY(guid) REFERENCES media,
  FOREIGN KEY(fId) REFERENCES filter,
  PRIMARY KEY(guid, fId)
);