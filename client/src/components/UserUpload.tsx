import { useState, useEffect } from "react";
import {
  uploadPending,
  uploadFile,
  updateDocumentStatus,
  listDocuments,
  type Document,
} from "../api";

export function UserUpload() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loadingDocs, setLoadingDocs] = useState(false);

  // This makes the page go blank sometimes - so leaving it out for now.
  useEffect(() => {
    loadDocuments();
  }, []);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const validTypes = [
        "application/pdf",
        "image/jpeg",
        "image/png",
        "image/gif",
        "image/webp",
      ];
      if (!validTypes.includes(file.type)) {
        setError("Please select a PDF or image file (JPEG, PNG, GIF, WebP)");
        return;
      }
      setSelectedFile(file);
      setError(null);
      setSuccess(false);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    setUploading(true);
    setError(null);

    try {
      const { upload_url, status_url } = await uploadPending(
        selectedFile.name,
        selectedFile.type,
      );

      const fileInfo = await uploadFile(upload_url, selectedFile);

      await updateDocumentStatus(
        status_url,
        "uploaded",
        fileInfo.file_size,
        fileInfo.check_sum,
      );

      setSuccess(true);
      setSelectedFile(null);

      loadDocuments();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setUploading(false);
    }
  };

  const loadDocuments = async () => {
    try {
      setLoadingDocs(true);
      const docs = await listDocuments();
      setDocuments(docs || []);
    } catch (err) {
      setDocuments([]);
      setError(err instanceof Error ? err.message : "Failed to load documents");
    } finally {
      setLoadingDocs(false);
    }
  };

  return (
    <main>
      <h1>Document Upload</h1>
      <p>Upload your documents (PDF or image)</p>

      <div style={{ marginBottom: "1rem" }}>
        <input
          type="file"
          accept=".pdf,image/*"
          onChange={handleFileChange}
          disabled={uploading}
        />
      </div>

      {selectedFile && (
        <div style={{ marginBottom: "1rem" }}>
          <p>
            Selected: {selectedFile.name} (
            {(selectedFile.size / 1024).toFixed(1)} KB)
          </p>
          <button onClick={handleUpload} disabled={uploading}>
            {uploading ? "Uploading..." : "Upload Document"}
          </button>
        </div>
      )}

      {error && <p style={{ color: "red" }}>{error}</p>}
      {success && (
        <p style={{ color: "green" }}>Document uploaded successfully!</p>
      )}

      <hr style={{ margin: "2rem 0" }} />

      <h2>My Documents</h2>
      <button onClick={loadDocuments} disabled={loadingDocs}>
        {loadingDocs ? "Loading..." : "Refresh"}
      </button>

      {documents.length === 0 ? (
        <p>No documents uploaded yet.</p>
      ) : (
        <div id="docs-grid" className="docs-grid">
          {documents.map((doc) => (
            <article className="docs-tile" key={doc.id}>
              <p className="docs-title">{doc.file_name}</p>
              <p className="docs-meta">Status: {doc.status}</p>
            </article>
          ))}
        </div>
      )}
    </main>
  );
}
