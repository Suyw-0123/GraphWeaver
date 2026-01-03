import { useState, useRef } from 'react';
import { api } from '../api/client';

interface UploadFormProps {
    onUploadSuccess: () => void;
    notebookId?: number;
}

export const UploadForm: React.FC<UploadFormProps> = ({ onUploadSuccess, notebookId }) => {
    const [isUploading, setIsUploading] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        console.log("UploadForm: File selected:", file?.name, "NotebookID:", notebookId);
        if (!file) return;

        if (!notebookId) {
            console.error("UploadForm: No notebook ID selected");
            alert("Please select a notebook first.");
            return;
        }

        setIsUploading(true);
        try {
            await api.uploadDocument(file, notebookId);
            onUploadSuccess();
            // Reset file input
            if (fileInputRef.current) {
                fileInputRef.current.value = '';
            }
        } catch (error: any) {
            console.error('Upload failed:', error);
            alert(`Failed to upload document: ${error.message || 'Unknown error'}`);
        } finally {
            setIsUploading(false);
        }
    };

    return (
        <div className="mb-4">
            <input
                type="file"
                ref={fileInputRef}
                onChange={handleFileChange}
                disabled={isUploading}
                accept=".pdf,.md,.txt"
                className="hidden"
                id="file-upload"
            />
            <label
                htmlFor="file-upload"
                className={`inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 cursor-pointer ${isUploading ? 'opacity-50 cursor-not-allowed' : ''}`}
            >
                {isUploading ? 'Uploading...' : 'Upload Document'}
            </label>
        </div>
    );
};
