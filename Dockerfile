FROM golang:1.23-alpine

WORKDIR /app

# Kopiere go.mod und go.sum, um Abhängigkeiten zu installieren
COPY go.mod go.sum ./

# Installiere MySQL-Treiber
RUN go mod tidy

# Kopiere den gesamten Code in den Container
COPY . .

# Baue die Go-Anwendung
RUN go build -o blackjack .
CMD ["./blackjack"]