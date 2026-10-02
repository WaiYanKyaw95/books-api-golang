import csv
import requests
import time

BASE_URL = "http://localhost:8080"

def get_all_books(url):
    books   = []
    page    = 1
    limit   = 10

    while True:
        # get response
        resp = requests.get(f"{url}/books", params={"page": page, "limit": limit})

        if resp.status_code == 429:
            print("rate limit hit")
            time.sleep(60)
            continue
        
        data = resp.json()

        # if no more records
        if not data:
            break

        for book in data:
            book["author"] = ", ".join(book["author"]) if book["author"] else ""
            book["subject"] = ", ".join(book["subject"]) if book["subject"] else ""

        # dump the data into a dict array
        books.extend(data)

        # pagination
        page += 1

    print(f"Total books fetched: {len(books)}")
    return books

def main():
    books = get_all_books(BASE_URL)
    
    with open("books.csv", "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=["id", "title", "author", "year", "subject", "description"])
        writer.writeheader()
        writer.writerows(books)

if __name__ == "__main__":
    main()