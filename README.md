# ddownloader

A simple, fast, and lightweight web-based file browser and downloader, specifically designed for NAS (Network Attached Storage) environments.

## 🚀 Motivation

The primary motivation behind **ddownloader** was to create a friction-less way to download files from a NAS. Most existing solutions are either too bloated, slow, or complex for simple file retrieval. 

This project aims to be:
- **Simple:** No complex configurations or heavy dependencies.
- **Fast:** Efficient directory scanning and size calculation.
- **Lightweight:** Minimal resource footprint, ideal for running on NAS hardware or low-power devices.

## ✨ Features

- **Recursive Directory Exploration:** Browse through your files with ease.
- **Real-time Size Calculation:** Automatically calculates the total size of directories.
- **Visual File Type Identification:** Icons and colors for different file formats (Videos, Images, PDFs, Archives).
- **Dockerized:** Ready to be deployed anywhere with a single command.
- **Responsive UI:** Clean and simple interface for both desktop and mobile.

## 🛠️ Installation & Usage

### Running Locally

Ensure you have [Go](https://go.dev/) installed (v1.25 or higher recommended).

1. Clone the repository.
2. Build the application:
   ```bash
   go build -o ddownloader ./cmd/ddownloader/main.go
   ```
3. Run the binary:
   ```bash
   ./ddownloader
   ```
4. Open your browser at `http://localhost:3000`.

### Running with Docker

1. Build the image:
   ```bash
   docker build -t ddownloader .
   ```
2. Run the container:
   ```bash
   docker run -p 3000:3000 -v /path/to/your/files:/data ddownloader
   ```
   *(Note: Ensure your application is configured to point to the correct volume path)*.

## 🧪 Testing

The project includes unit tests for core logic and utilities. To run them:

```bash
go test ./...
```
