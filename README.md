# Merkle Tree File Verification Backend

A backend server for demonstrating the functionality of a Merkle tree for a file storage and integrity verification system. Built using Go and PostgreSQL. The server exposes endpoints for uploading, downloading and generating Merkle proofs to ensure file integrity.

This backend is written to be used in conjuction with the [Merkle Tree File Verification Client](https://github.com/CaelRowley/merkle-tree-file-verification-client).

## Prerequisites

- [Docker](https://docs.docker.com/desktop/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Environment Variables

- **DB_URL**
  - This specifies the URL of the DB.
  - Defaults to `postgresql://admin:admin@localhost:5432` (points to the default URL of the DB from this service).
  - If you want to connect to a deployed DB, set `DB_URL` to the URL of your DB (must be a PostgreSQL DB):
    ```bash
    export DB_URL=postgres://db_user:db_password@db_postgres.db.com/db_name
    ```

## Getting Started

After cloning the repository, you can run the backend server and locally using Docker Compose.

**Note** The backend runs on port `8080` and the DB on port `5432` so make sure those ports are free if running locally.

**Step 1:** Build the Docker Images (if needed)

```bash
docker-compose -f docker-compose.yml build
```

**Step 2:** Run the server and DB

```bash
docker-compose -f docker-compose.yml up
```

## Backend Routes

The backend server exposes the following routes:

- **POST** `/files/upload-batch/{id}`: Upload files to the database.
  - queryParams add `?batch-complete=true` to signal the last batch
- **POST** `/files/delete-all`: Delete all files from the database.
- **GET** `/files/download/{id}`: Download a file from the database.
- **GET** `/files/get-proof/{id}`: Get a Merkle proof for a file from the database.
- **POST** `/files/corrupt-file/{id}`: Simulate file corruption in the database by modifying a file and not the hash.

## Database Table Structure

The backend server utilizes a PostgreSQL database to store file metadata and content. Below is the structure of the `files` table used by the server:

| Column        | Type      | Description                                                                             |
|---------------|-----------|-----------------------------------------------------------------------------------------|
| id            | integer   | Sequence-generated integer used as the primary key for files.                           |
| batch_id      | uuid      | Universally unique identifier (UUID) used to identify which batch a file belongs to.    |
| name          | text      | The name of the file.                                                                   |
| file          | bytea     | The content of the file stored as binary.                                               |
| original_hash | bytea     | The SHA256 hash of the file stored as binary.                                           |

### Purpose of Each Column

- **id:** Primary key used for unique identification of files.
- **batch_id:** Files are uploaded in batches and this batch_id is used to group the files together in a single batch.
- **name:** Original filename of the uploaded file.
- **file:** Binary content (file data) of the uploaded file.
- **original_hash:** This plays a crucial role in generating Merkle proofs for file integrity verification. By using the hash of each file's content, the server or client can detect any modifications or tampering attempts on the file data. This field allows the server to simulate malicious activity (e.g., file corruption) by modifying the file content while keeping a reference to original unmodified file hash.

