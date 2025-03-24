FROM golang:1.23.1
COPY starGo /app/starGo
WORKDIR /app
CMD ["/app/starGo"]