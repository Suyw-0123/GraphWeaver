export interface Document {
    id: number;
    filename: string;
    file_path: string;
    mime_type: string;
    file_size: number;
    status: 'pending' | 'processing' | 'completed' | 'failed';
    error_message?: string;
    summary?: string;
    created_at: string;
    updated_at: string;
    is_deleted: boolean;
    notebook_id?: number;
}

export interface Notebook {
    id: number;
    title: string;
    description: string;
    created_at: string;
    updated_at: string;
}

export interface CreateDocumentRequest {
    filename: string;
    file_path: string;
    mime_type: string;
    file_size: number;
}
