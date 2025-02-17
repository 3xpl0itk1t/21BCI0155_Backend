## TradeMarkia API Documentation

This document provides an overview of the API endpoints available for the TradeMarkia backend.


[<img src="https://run.pstmn.io/button.svg" alt="Run In Postman" style="width: 128px; height: 32px;">](https://god.gw.postman.com/run-collection/25976453-23f6155d-c6ac-4e9a-abf5-ea2c10d02565?action=collection%2Ffork&source=rip_markdown&collection-url=entityId%3D25976453-23f6155d-c6ac-4e9a-abf5-ea2c10d02565%26entityType%3Dcollection%26workspaceId%3D2273b065-77b7-4c9e-bc25-bcf0749ee27d)

### Base URL

The base URL for all endpoints is:

```
http://13.51.204.39:8000
```

### Endpoints

#### Register

Registers a new user.

**Method:** POST

**Endpoint:** /register

**Request Body (JSON):**

```json
{
  "email": "prabhavmishra7@gmail.com",
  "username": "prxbhav",
  "password": "prabhavg"
}
```

**Example using curl:**

```bash
curl --location --request POST 'http://13.51.204.39:8000/register' \
--header 'Content-Type: application/json' \
--data-raw '{
  "email": "prabhavmishra7@gmail.com",
  "username": "prxbhav",
  "password": "prabhavg"
}'
```

#### Login

Log in and get a JWT token.

**Method:** POST

**Endpoint:** /login

**Request Body (JSON):**

```json
{
  "email": "prabhavmishra7@gmail.com",
  "password": "prabhavg"
}
```

**Example using curl:**

```bash
curl --location --request POST 'http://13.51.204.39:8000/login' \
--header 'Content-Type: application/json' \
--data-raw '{
  "email": "prabhavmishra7@gmail.com",
  "password": "prabhavg"
}'
```

#### File Upload

Upload a file.

**Method:** POST

**Endpoint:** /upload

**Request Headers:**

* Authorization: Bearer your-jwt-token

**Request Body:**

* Form Data:
    * file: [Upload a file]

**Example using curl:**

```bash
curl --location --request POST 'http://13.51.204.39:8000/upload' \
--header 'Authorization: Bearer your-jwt-token' \
--form 'file=@/path/to/your/file'
```

#### Files

Retrieve metadata for all files uploaded by the user.

**Method:** GET

**Endpoint:** /files

**Request Headers:**

* Authorization: Bearer your-jwt-token

**Example using curl:**

```bash
curl --location --request GET 'http://13.51.204.39:8000/files' \
--header 'Authorization: Bearer your-jwt-token'
```

#### Share Files

Get a public link to share a file.

**Method:** GET

**Endpoint:** /share/{file_id}

**Request Headers:**

* Authorization: Bearer your-jwt-token

**Example using curl:**

```bash
curl --location --request GET 'http://13.51.204.39:8000/share/your-file-id' \
--header 'Authorization: Bearer your-jwt-token'
```

#### Search Files

Search for files by name, date, limit, and offset.

**Method:** GET

**Endpoint:** /search

**Query Parameters:**

* name: Name of the file
* date: Date of the file
* limit: Number of results to return
* offset: Number of results to skip

**Request Headers:**

* Authorization: Bearer your-jwt-token

**Example using curl:**

```bash
curl --location --request GET 'http://13.51.204.39:8000/search?name=check.jpg&date=2024-09-15%2002:57:15.494094&limit=10&offset=0' \
--header 'Authorization: Bearer your-jwt-token'
```

### Running the Project

**Start the Application:**

```bash
docker-compose up --build
```

**Use Postman or curl to send requests to my API.**

### License

This project is licensed under the MIT License - see the LICENSE file for details.
