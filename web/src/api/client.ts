import { Document, CreateDocumentRequest, Notebook } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

export const api = {
    // Notebooks
    async listNotebooks(): Promise<Notebook[]> {
        const response = await fetch(`${API_BASE_URL}/notebooks`);
        if (!response.ok) throw new Error('Failed to fetch notebooks');
        return response.json();
    },

    async createNotebook(title: string, description: string): Promise<Notebook> {
        const response = await fetch(`${API_BASE_URL}/notebooks`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, description }),
        });
        if (!response.ok) throw new Error('Failed to create notebook');
        return response.json();
    },

    async getNotebook(id: number): Promise<Notebook> {
        const response = await fetch(`${API_BASE_URL}/notebooks/${id}`);
        if (!response.ok) throw new Error('Failed to fetch notebook');
        return response.json();
    },

    async deleteNotebook(id: number): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/notebooks/${id}`, { method: 'DELETE' });
        if (!response.ok) throw new Error('Failed to delete notebook');
    },

    // Documents
    async listDocuments(notebookId?: number): Promise<Document[]> {
        let url = `${API_BASE_URL}/documents`;
        if (notebookId) {
            url += `?notebook_id=${notebookId}`;
        }
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Failed to fetch documents');
        }
        return response.json();
    },

    async createDocument(data: CreateDocumentRequest): Promise<Document> {
        const response = await fetch(`${API_BASE_URL}/documents`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });
        if (!response.ok) {
            throw new Error('Failed to create document');
        }
        return response.json();
    },

    async uploadDocument(file: File, notebookId: number): Promise<Document> {
        console.log("API: Uploading document", file.name, "to notebook", notebookId);
        const formData = new FormData();
        formData.append('file', file);
        formData.append('notebook_id', notebookId.toString());

        const response = await fetch(`${API_BASE_URL}/documents/upload`, {
            method: 'POST',
            body: formData,
        });

        console.log("API: Upload response status:", response.status);

        if (!response.ok) {
            const text = await response.text();
            console.error("API: Upload failed response:", text);
            throw new Error(`Failed to upload document: ${text}`);
        }
        return response.json();
    },

    async getDocumentGraph(documentId: number): Promise<any> {
        const response = await fetch(`${API_BASE_URL}/documents/${documentId}/graph`);
        if (!response.ok) {
            throw new Error('Failed to fetch graph data');
        }
        return response.json();
    },

    async chat(notebookId: number, query: string): Promise<string> {
        const response = await fetch(`${API_BASE_URL}/notebooks/${notebookId}/chat`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ query }),
        });
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || 'Failed to send message');
        }
        const data = await response.json();
        return data.answer;
    }
};
