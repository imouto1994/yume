CREATE TABLE LIBRARY (
  ID INTEGER PRIMARY KEY,
  NAME TEXT NOT NULL,
  ROOT TEXT NOT NULL
);

CREATE TABLE TITLE (
  ID INTEGER PRIMARY KEY,
  NAME TEXT NOT NULL,
  URL TEXT NOT NULL,
  CREATED_AT TEXT NOT NULL,
  UPDATED_AT TEXT NOT NULL,
  COVER_WIDTH INTEGER NOT NULL,
  COVER_HEIGHT INTEGER NOT NULL,
  BOOK_COUNT INTEGER NOT NULL,
  UNCENSORED INTEGER NOT NULL,
  WAIFU2X INTEGER NOT NULL,
  LANGS TEXT NOT NULL,
  LIBRARY_ID INTEGER NOT NULL,
  FOREIGN KEY (LIBRARY_ID) REFERENCES LIBRARY (ID)
);
CREATE INDEX idx__title__library_id on TITLE (LIBRARY_ID);

CREATE TABLE BOOK (
  ID INTEGER PRIMARY KEY,
  NAME TEXT NOT NULL,
  URL TEXT NOT NULL,
  CREATED_AT TEXT NOT NULL,
  UPDATED_AT TEXT NOT NULL,
  PREVIEW_URL TEXT,
  PREVIEW_UPDATED_AT TEXT,
  PAGE_COUNT INTEGER NOT NULL,
  TITLE_ID INTEGER NOT NULL,
  LIBRARY_ID INTEGER NOT NULL,
  FOREIGN KEY (LIBRARY_ID) REFERENCES LIBRARY (ID),
  FOREIGN KEY (TITLE_ID) REFERENCES TITLE (ID)
);
CREATE INDEX idx__book__title_id on BOOK (TITLE_ID);
CREATE INDEX idx__book__library_id on BOOK (LIBRARY_ID);

CREATE TABLE PAGE (
  NUMBER INTEGER NOT NULL,
  FILE_INDEX INTEGER NOT NULL,
  WIDTH INTEGER NOT NULL,
  HEIGHT INTEGER NOT NULL,
  FAVORITE INTEGER NOT NULL,
  BOOK_ID INTEGER NOT NULL,
  TITLE_ID INTEGER NOT NULL,
  LIBRARY_ID INTEGER NOT NULL,
  PRIMARY KEY (BOOK_ID, NUMBER),
  FOREIGN KEY (BOOK_ID) REFERENCES BOOK (ID),
  FOREIGN KEY (LIBRARY_ID) REFERENCES LIBRARY (ID),
  FOREIGN KEY (TITLE_ID) REFERENCES TITLE (ID)
);
CREATE INDEX idx__page__book_id on PAGE (BOOK_ID);
CREATE INDEX idx__page__title_id on PAGE (TITLE_ID);
CREATE INDEX idx__page__library_id on PAGE (LIBRARY_ID);

CREATE TABLE PREVIEW (
  NUMBER INTEGER NOT NULL,
  FILE_INDEX INTEGER NOT NULL,
  BOOK_ID INTEGER NOT NULL,
  TITLE_ID INTEGER NOT NULL,
  LIBRARY_ID INTEGER NOT NULL,
  PRIMARY KEY (BOOK_ID, NUMBER),
  FOREIGN KEY (BOOK_ID) REFERENCES BOOK (ID),
  FOREIGN KEY (LIBRARY_ID) REFERENCES LIBRARY (ID),
  FOREIGN KEY (TITLE_ID) REFERENCES TITLE (ID)
);
CREATE INDEX idx__preview__book_id on PREVIEW (BOOK_ID);
CREATE INDEX idx__preview__title_id on PREVIEW (TITLE_ID);
CREATE INDEX idx__preview__library_id on PREVIEW (LIBRARY_ID);