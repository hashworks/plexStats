-- name: update-server
UPDATE server SET `name` = ? WHERE uuid = ?;

-- name: insert-server
INSERT INTO server(uuid, `name`) VALUES(?, ?);

-- name: update-account
UPDATE account SET `name`= ?, thumbnail = ? WHERE plexnumber = ?;

-- name: insert-account
INSERT INTO account(plexNumber, `name`, thumbnail) VALUES(?, ?, ?);

-- name: select-address-id-by-ip
SELECT aId FROM address WHERE ip = ? LIMIT 1;

-- name: insert-address
INSERT INTO address(ip) VALUES(?);

-- name: update-client
UPDATE client SET `name`= ? WHERE uuid = ?;

-- name: insert-client
INSERT INTO client(uuid, `name`) VALUES(?, ?);

-- name: update-media
UPDATE media SET
  type = ?,
  subtype = ?,

  key = ?,
  parentKey = ?,
  grandparentKey = ?,
  primaryExtraKey = ?,

  title = ?,
  titleSort = ?,
  parentTitle = ?,
  grandparentTitle = ?,

  summary = ?,
  duration = ?,

  thumb = ?,
  parentThumb = ?,
  grandparentThumb = ?,

  grandparentTheme = ?,
  grandparentRatingKey = ?,

  art = ?,
  grandparentArt = ?,

  `index` = ?,
  parentIndex = ?,

  studio = ?,
  tagline = ?,
  chapterSource = ?,

  librarySectionID = ?,
  librarySectionKey = ?,
  librarySectionType = ?,

  webRating = ?,
  userRating = ?,
  audienceRating = ?,
  contentRating = ?,
  ratingImage = ?,
  viewCount = ?,

  releaseYear = ?,
  dateOriginal = ?,
  dateAdded = ?,
  dateUpdated = ?
  WHERE guid = ?;

-- name: insert-media
INSERT INTO media(
  guid,

  type,
  subtype,

  key,
  parentKey,
  grandparentKey,
  primaryExtraKey,

  title,
  titleSort,
  parentTitle,
  grandparentTitle,

  summary,
  duration,

  thumb,
  parentThumb,
  grandparentThumb,

  grandparentTheme,
  grandparentRatingKey,

  art,
  grandparentArt,

  `index`,
  parentIndex,

  studio,
  tagline,
  chapterSource,

  librarySectionID,
  librarySectionKey,
  librarySectionType,

  webRating,
  userRating,
  audienceRating,
  contentRating,
  ratingImage,
  viewCount,

  releaseYear,
  dateOriginal,
  dateAdded,
  dateUpdated) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: insert-event
INSERT INTO event(date, type, rating, local, owned, accountNumber, sUUID, cUUID, mGUID, aId) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: update-filter
UPDATE filter SET tag = ?, filter = ?, role = ?, thumb = ?, count = ? WHERE fId = ?;

-- name: insert-filter
INSERT INTO filter(fId, tag, filter, role, thumb, count) VALUES(?, ?, ?, ?, ?, ?);