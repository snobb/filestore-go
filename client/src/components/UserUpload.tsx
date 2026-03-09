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
    <div style={{ padding: "2rem", maxWidth: "600px", margin: "0 auto" }}>
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
        <div
          style={{
            display: "table",
            width: "100%",
            borderCollapse: "collapse",
          }}
        >
          <div style={{ display: "table-header-group" }}>
            <div
              style={{
                display: "table-row",
                borderBottom: "1px solid #ccc",
                fontWeight: "bold",
              }}
            >
              <div style={{ display: "table-cell", padding: "8px" }}>
                File Name
              </div>
              <div style={{ display: "table-cell", padding: "8px" }}>
                Status
              </div>
            </div>
          </div>
          <div style={{ display: "table-row-group" }}>
            {documents.map((doc) => (
              <div
                key={doc.id}
                style={{ display: "table-row", borderBottom: "1px solid #eee" }}
              >
                <div style={{ display: "table-cell", padding: "8px" }}>
                  {doc.file_name}
                </div>
                <div style={{ display: "table-cell", padding: "8px" }}>
                  <strong>{doc.status}</strong>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
