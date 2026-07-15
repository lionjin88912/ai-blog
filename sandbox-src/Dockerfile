FROM golang:1.22-alpine AS builder

# gcc/musl-dev needed for creack/pty CGO fallback (if enabled)
RUN apk add --no-cache gcc musl-dev

WORKDIR /src

# Copy source
COPY . .

# Resolve dependencies
RUN go mod tidy
RUN go mod download

# Verify code compiles
RUN go vet ./...

# Build all targets
RUN mkdir -p /out
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /out/ai-sandbox-linux-amd64 .
RUN CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-s -w" -o /out/ai-sandbox-darwin-amd64 .
RUN CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "-s -w" -o /out/ai-sandbox-darwin-arm64 .
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o /out/ai-sandbox-windows-amd64.exe .
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o /out/pwsh.exe ./cmd/pwsh-shim
RUN cp ai-sandbox-windows-amd64.bat /out/
RUN cp bash.bat /out/

FROM alpine:3.20
COPY --from=builder /out/ /out/
CMD ["cp", "-r", "/out/.", "/dist/"]
