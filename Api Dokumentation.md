# Dokumentation:
## Allgemein
Es wird ein Port benötig, auf dem die Software erreichbar ist.
## Mögliche URLs
GET / -> Neue Weiterleitung erstellen
PUT / -> Weiterleitung erstellen
GET /:key -> Weiterleitung nutzen
DELETE /:key -> Weiterleitung löschen
## PUT /
Fügt eine URL zur Datenbank hinzu
###Request Body:
```
{
    url: "google.com",
    mode: 1
}
```
### Response Body:
```
{
    key: "abc123"
}
