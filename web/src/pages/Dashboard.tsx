import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { Notebook } from '../types';

export const Dashboard: React.FC = () => {
    const [notebooks, setNotebooks] = useState<Notebook[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [newTitle, setNewTitle] = useState('');
    const [newDesc, setNewDesc] = useState('');

    const fetchNotebooks = async () => {
        try {
            setLoading(true);
            const data = await api.listNotebooks();
            setNotebooks(data || []);
            setError(null);
        } catch (error) {
            console.error(error);
            setError("Failed to load notebooks. Please check if the backend server is running.");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchNotebooks();
    }, []);

    const handleCreate = async () => {
        if (!newTitle) return;
        try {
            await api.createNotebook(newTitle, newDesc);
            setShowCreateModal(false);
            setNewTitle('');
            setNewDesc('');
            fetchNotebooks();
        } catch (error) {
            console.error(error);
            alert('Failed to create notebook');
        }
    };

    const handleDelete = async (id: number, e: React.MouseEvent) => {
        e.preventDefault(); // Prevent navigation
        if (!confirm('Are you sure? This will delete all documents in this notebook.')) return;
        try {
            await api.deleteNotebook(id);
            fetchNotebooks();
        } catch (error) {
            console.error(error);
            alert('Failed to delete notebook');
        }
    };

    if (loading) return <div className="p-8 text-center text-gray-400">Loading notebooks...</div>;
    if (error) return (
        <div className="p-8 text-center">
            <div className="text-red-400 mb-4">{error}</div>
            <button onClick={fetchNotebooks} className="text-indigo-400 hover:underline">Retry</button>
        </div>
    );

    return (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="flex justify-between items-center mb-8">
                <h1 className="text-3xl font-bold text-gray-100">Notebooks</h1>
                <button
                    onClick={() => setShowCreateModal(true)}
                    className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700 transition-colors"
                >
                    New Notebook
                </button>
            </div>

            {notebooks.length === 0 ? (
                <div className="text-center py-12 bg-gray-800 rounded-lg shadow border border-gray-700">
                    <p className="text-gray-400 mb-4">No notebooks found.</p>
                    <button
                        onClick={() => setShowCreateModal(true)}
                        className="text-indigo-400 hover:text-indigo-300 font-medium"
                    >
                        Create your first notebook
                    </button>
                </div>
            ) : (
                <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                    {notebooks.map((nb) => (
                        <Link
                            key={nb.id}
                            to={`/notebook/${nb.id}`}
                            className="block bg-gray-800 overflow-hidden shadow rounded-lg hover:shadow-md transition-all border border-gray-700 hover:border-gray-600"
                        >
                            <div className="px-4 py-5 sm:p-6">
                                <div className="flex justify-between items-start">
                                    <h3 className="text-lg font-medium text-gray-100 truncate">{nb.title}</h3>
                                    <button
                                        onClick={(e) => handleDelete(nb.id, e)}
                                        className="text-gray-500 hover:text-red-400 transition-colors"
                                    >
                                        <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                        </svg>
                                    </button>
                                </div>
                                <p className="mt-1 text-sm text-gray-400 line-clamp-2">{nb.description}</p>
                                <div className="mt-4 flex items-center text-xs text-gray-500">
                                    <span>Created {new Date(nb.created_at).toLocaleDateString()}</span>
                                </div>
                            </div>
                        </Link>
                    ))}
                </div>
            )}

            {showCreateModal && (
                <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center p-4 z-50 backdrop-blur-sm">
                    <div className="bg-gray-800 rounded-lg max-w-md w-full p-6 border border-gray-700 shadow-xl">
                        <h2 className="text-xl font-bold mb-4 text-gray-100">Create New Notebook</h2>
                        <div className="mb-4">
                            <label className="block text-sm font-medium text-gray-300 mb-1">Title</label>
                            <input
                                type="text"
                                value={newTitle}
                                onChange={(e) => setNewTitle(e.target.value)}
                                className="w-full border border-gray-600 bg-gray-700 rounded-md p-2 text-white focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                                placeholder="My Research"
                            />
                        </div>
                        <div className="mb-6">
                            <label className="block text-sm font-medium text-gray-300 mb-1">Description</label>
                            <textarea
                                value={newDesc}
                                onChange={(e) => setNewDesc(e.target.value)}
                                className="w-full border border-gray-600 bg-gray-700 rounded-md p-2 text-white focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                                rows={3}
                                placeholder="Notes about..."
                            />
                        </div>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setShowCreateModal(false)}
                                className="px-4 py-2 text-gray-300 hover:text-white transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleCreate}
                                className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700 transition-colors"
                            >
                                Create
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};
