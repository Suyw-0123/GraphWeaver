import { Document } from '../types';

interface DocumentListProps {
    documents: Document[];
    loading: boolean;
    error: string | null;
}

export const DocumentList: React.FC<DocumentListProps> = ({ documents, loading, error }) => {
    if (loading && documents.length === 0) return <div className="text-center p-4">Loading...</div>;
    if (error) return <div className="text-red-500 p-4">{error}</div>;

    return (
        <div className="bg-gray-800 shadow rounded-lg overflow-hidden">
            <div className="px-4 py-5 sm:px-6 border-b border-gray-700">
                <h3 className="text-lg leading-6 font-medium text-gray-100">Document</h3>
            </div>
            <div className="">
                <ul className="divide-y divide-gray-700">
                    {documents.map((doc) => (
                        <li key={doc.id} className="px-4 py-4 sm:px-6 hover:bg-gray-700 transition-colors">
                            <div className="flex items-center justify-between">
                                <div className="text-sm font-medium text-indigo-400 truncate">
                                    {doc.filename}
                                </div>
                                <div className="ml-2 flex-shrink-0 flex">
                                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                                        ${doc.status === 'completed' ? 'bg-green-900 text-green-200' : 
                                          doc.status === 'failed' ? 'bg-red-900 text-red-200' : 
                                          'bg-yellow-900 text-yellow-200'}`}>
                                        {doc.status}
                                    </span>
                                </div>
                            </div>
                            <div className="mt-2 sm:flex sm:justify-between">
                                <div className="sm:flex">
                                    <p className="flex items-center text-sm text-gray-400">
                                        {doc.mime_type} • {(doc.file_size / 1024).toFixed(2)} KB
                                    </p>
                                </div>
                                <div className="mt-2 flex items-center text-sm text-gray-400 sm:mt-0">
                                    <p>
                                        Uploaded on {new Date(doc.created_at).toLocaleDateString()}
                                    </p>
                                </div>
                            </div>
                            {doc.status === 'failed' && doc.error_message && (
                                <div className="mt-2 p-2 bg-red-900/50 rounded-md text-xs text-red-300 border border-red-800">
                                    Error: {doc.error_message}
                                </div>
                            )}
                            {doc.summary && (
                                <div className="mt-4 p-3 bg-gray-700/50 rounded-md border border-gray-600">
                                    <p className="text-sm font-medium text-gray-300 mb-1">AI Summary:</p>
                                    <p className="text-sm text-gray-400 whitespace-pre-wrap">{doc.summary}</p>
                                </div>
                            )}
                        </li>
                    ))}
                    {documents.length === 0 && (
                        <li className="px-4 py-8 text-center text-gray-500">
                            No document found. Upload one to get started.
                        </li>
                    )}
                </ul>
            </div>
        </div>
    );
};
