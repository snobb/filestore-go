## Starting a project

```bash
docker compose up --build -d
```
This should build both frontend and backend system and spin up the database and
the backend server.

## Requirements

### Functional requirements:

- Users should be able to upload a document as either a PDF or a standard image
  format.
- The system should store the document and track its status.
- Solicitors should be able to see a list of uploads and their status.

### Product and technical Requirements
- Frontend: Provide a simple interface that allows a user to select and upload
  a document, e.g a photo of their passport.
- Backend: Create an API to receive the file, validate it, and facilitate
  storage.
- Data Persistence: The system must track the status of the document (e.g.,
  Pending, Verified, Rejected) and associated metadata.
- Administration: Provide a way for a solicitor to view a list of uploaded
  documents and their current status.

## APIs

### REST APIs

**POST /api/documents**
- Accepts the file and metadata.

Request:

```json
    {
        "file_name": "passport.pdf",
        "content_type": "application/pdf"
    }
```
Response:
```json
{
  "id": "uuid-123",
  "upload_url": "/file_store/uploads/uuid-123",
  "status_url": "/api/documents/uuid-123/status"
}
```

**PATCH /api/documents/{id}/status**
- Updates the status (Pending → Verified).*

Request:
```json
{
  "status":"uploaded",
  "file_size":123,
  "checksum":"abcdef123"
}
```
Response:
```json
{
  "status": "uploaded"
  "uploaded_at": "2022-01-01T00:00:00Z"
}
```

**GET /api/documents**
- Returns a list of files for the given user_id

Response:
```json
[
  {
    "id": "uuid-123",
    "user_id": "user-uuid-345",
    "file_name": "passport.pdf",
    "content_type": "application/pdf",
    "status": "Pending"
  },
  ...
]
```

**GET /api/documents/{id}**
- Provide Document metadata

Response:
```json
{
  "id": "uuid-123",
  "user_id": "user-uuid-345",
  "file_name": "passport.pdf",
  "content_type": "application/pdf",
  "status": "Pending"
}
```

**GET /file_store//downloads/path/to/file/**
- Download the file

**POST /file_store/uploads/path/to/file**

- Mock S3 file store. Document metadata will contain the pre-signed URI
  stream the contents of the file to the given url.
  In order to save the time I do not endevour to mimic the S3 api behaviour.

### DB schema
Database schema:


```sql
-- for simplicity and being mindful of time - no auth in this project
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE doc_status AS ENUM ('pending', 'uploaded', 'verified', 'rejected');

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    file_name TEXT UNIQUE NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    content_type TEXT NOT NULL,
    checksum TEXT,
    status doc_status DEFAULT 'pending',
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- optimistic idempotency control.
ALTER TABLE documents ADD CONSTRAINT unique_user_filename
UNIQUE (user_id, file_name);

ALTER TABLE documents ADD CONSTRAINT fk_documents_user
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Demo user
INSERT INTO users (id, email, username, hashed_password)
VALUES ('00000000-0000-0000-0000-000000000000', 'test@snobb.org', 'DemoClient', 'nopassword');
```

## Architecture

### Design choices

#### Split the file store and metadata store.
- I'm using the postgresql database as the metadata storage (as suggested in
  the template project - but it would be my choice anyway).
- Use mocked FS file storage for this project.
- On the server side - I've split the code in packages that separate documents,
  mock filestore and auth.

#### Data Model
- Here I've tied files to a userID. But I normally it would be better to add a
  kind of caseID or groupID to group files together. This will allow for a move
  flexible authorization model (Eg. multiple users have access to the same files)


### Project structure:

#### Changes
- moved db the provided db/compose.yaml -> docker-compose.yaml file to the root to serve for both db
  and server.
- changed the go project structure as per https://github.com/golang-standards/project-layout
  The `main.go` is moved to `cmd/server/main.go`
- the react vite project structure is not changed.
- a new folder `filestore` is created automatically as a file storage.

#### Frontend
2 different screens for solicitors and users:
- Solicitors should be able to see a list of uploads and their status (and download them).
- Users should be able to upload a document as either a PDF or a standard image format.

#### Backend
packages:
- document - file handler
- filestore - File storage layer (s3 moc)
- auth - authentication mock (middleware that sets userid in the context
  pretending to have parsed JWT).

The test coverage is not complete given the time constraints. Most tests were
vibe coded using the mocks that I've generated with moq tool.

## Assumptions
Given the time constraints, the following assumptions were made:
- Even thought the verified and rejected mentioned in the descriptions. I do
  not understand the workflow, so I only add pending and uploaded states.
  - Pending - upload initiated by user but file hasn't been uploaded yet.
  - Uploaded - file has been uploaded.

- No HTTPS or TLS is used.

- No authentication or authorization. I've made a provisioning for
  authentication in the future but here we'll go with a hard coded user ID and
  a middleware simulating JWT parsing.

- I'll assume that there is no concept of folders or subfolders. Each user has
  a single folder to store files. I'll use the userID as the folder name.

- Mocked file store. Typically a file store like S3 or GCS file store is used.
  So I'm using a local FS for this project but the logic flow accommodates for
  the S3 use case.

- No file post-upload processing - I'd assume there would be a separate service
  that does the file verification (eg. antivirus scan, OCR, etc) and stores the
  results.
  The way I'd do it is - I'd trigger an event on upload (eg. kafka message) for
  another service to process the file and update the status via the status API
  endpoint in the database (eg. uploaded -> verified).

## Problems and potential improvements
- No pagination for the list of document endpoints. Typically I'd add
  pagination which typically involves passing the pagination parameters and
  updating the select queries.

- I specifically allow multiple copies of the same file by prefixing fileID to
  avoid overwriting at the filestore. Idempotency is provided by means of the
  React, disabling the upload button on successful upload.
  This has a problem thought - if the upload takes too long - it's possible to
  create the upload button multiple times and create multiple records.
  So there should be a way to prevent it. Here are possible options how to do that:
  - add a unique constraint on file_name and userID, which works for the poc
    but there are better options.
  - add a correlationID or idempotencyToken that would be generated by react
    and stored in the database. Then the database can be used for optimistic
    locking using the field as a constraint.

- I could add another abstraction for document service that would do the all
  the logic of saving to database and dealing with the file storage. That would
  make the HTTP handlers part of the code more readable and testable.

- There are some input validation

- Test coverage is not complete and can definitely be improved.

## Issues
- Sometimes the react page ends up blank (especially on first start). I think I
  have fixed it, but it's not 100% sure.
