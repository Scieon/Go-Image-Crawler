# Go-Image-Crawler

```
docker build -t crawler_test -f Dockerfile ./
docker run -it -p 8080:8080 crawler_test
```

Example request:

```
curl -X POST 'http://localhost:8080' -H "Content-Type: application/json" --data '{
        "threads": 1,
        "urls": ["http://golang.com", "https://www.golang-book.com/"]
}'
```
