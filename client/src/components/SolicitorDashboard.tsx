import { useState, useEffect } from "react";
import { listDocuments, getDownloadUrl, type Document } from "../api";

export function SolicitorDashboard() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadDocuments = async () => {
    setLoading(true);
    setError(null);
    try {
      const docs = await listDocuments();
      setDocuments(docs || []);
    } catch (err) {
      setDocuments([]);
      setError(err instanceof Error ? err.message : "Failed to load documents");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDocuments();
  }, []);

  const getStatusColor = (status: string) => {
    switch (status) {
      case "verified":
        return "green";
      case "uploaded":
        return "blue";
      default:
        return "gray";
    }
  };

  return (
    <div style={{ padding: "2rem", maxWidth: "800px", margin: "0 auto" }}>
      <h1>Solicitor Dashboard</h1>
      <p>Review uploaded documents</p>

      <button
        onClick={loadDocuments}
        disabled={loading}
        style={{ marginBottom: "1rem" }}
      >
        {loading ? "Loading..." : "Refresh"}
      </button>

      {error && <p style={{ color: "red" }}>{error}</p>}

      {documents.length === 0 ? (
        <p>No documents uploaded yet.</p>
      ) : (
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead>
            <tr style={{ textAlign: "left", borderBottom: "1px solid #ccc" }}>
              <th>File Name</th>
              <th>Status</th>
              <th>Uploaded</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {documents.map((doc) => (
              <tr key={doc.id} style={{ borderBottom: "1px solid #eee" }}>
                <td>{doc.file_name}</td>
                <td>
                  <span
                    style={{
                      color: getStatusColor(doc.status),
                      fontWeight: "bold",
                    }}
                  >
                    {doc.status}
                  </span>
                </td>
                <td>{new Date(doc.uploaded_at).toLocaleDateString()}</td>
                <td>
                  <a
                    href={getDownloadUrl(doc.file_path)}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Download
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
