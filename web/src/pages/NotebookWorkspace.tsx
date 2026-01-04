import { useEffect, useState, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import ForceGraph2D from 'react-force-graph-2d';
import ReactMarkdown from 'react-markdown';
import { api } from '../api/client';
import { Notebook, Document } from '../types';
import { DocumentList } from '../components/DocumentList';
import { UploadForm } from '../components/UploadForm';

export const NotebookWorkspace: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const notebookId = parseInt(id || '0');
    const [notebook, setNotebook] = useState<Notebook | null>(null);
    const [documents, setDocuments] = useState<Document[]>([]);
    const [loading, setLoading] = useState(true);
    const [docsLoading, setDocsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [docsError, setDocsError] = useState<string | null>(null);
    const [graphData, setGraphData] = useState<any>(null);
    const graphContainerRef = useRef<HTMLDivElement>(null);
    const [graphDimensions, setGraphDimensions] = useState({ width: 0, height: 0 });

    // Chat State
    const [messages, setMessages] = useState<{ role: 'user' | 'assistant'; content: string }[]>([
        { role: 'assistant', content: 'Hello! I am your AI assistant for this notebook. Upload a document to get started.' }
    ]);
    const [input, setInput] = useState('');
    const [isChatting, setIsChatting] = useState(false);
    const messagesEndRef = useRef<HTMLDivElement>(null);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };

    useEffect(() => {
        scrollToBottom();
    }, [messages]);

    const handleSendMessage = async (e?: React.FormEvent) => {
        e?.preventDefault();
        if (!input.trim() || isChatting) return;

        const userMessage = input.trim();
        setMessages(prev => [...prev, { role: 'user', content: userMessage }]);
        setInput('');
        setIsChatting(true);

        try {
            const answer = await api.chat(notebookId, userMessage);
            setMessages(prev => [...prev, { role: 'assistant', content: answer }]);
        } catch (error) {
            console.error("Chat error:", error);
            setMessages(prev => [...prev, { role: 'assistant', content: "Sorry, I encountered an error processing your request." }]);
        } finally {
            setIsChatting(false);
        }
    };

    useEffect(() => {
        const updateDimensions = () => {
            if (graphContainerRef.current) {
                setGraphDimensions({
                    width: graphContainerRef.current.offsetWidth,
                    height: graphContainerRef.current.offsetHeight
                });
            }
        };

        window.addEventListener('resize', updateDimensions);
        updateDimensions();
        
        // Small delay to ensure layout is done
        setTimeout(updateDimensions, 100);

        return () => window.removeEventListener('resize', updateDimensions);
    }, [graphContainerRef.current, graphData]); // Update when data changes too, as layout might change

    const fetchDocuments = async (isPolling = false) => {
        try {
            if (!isPolling) setDocsLoading(true);
            const docs = await api.listDocuments(notebookId);
            setDocuments(docs || []);
            setDocsError(null);
        } catch (err) {
            console.error(err);
            if (!isPolling) setDocsError('Failed to load documents');
        } finally {
            if (!isPolling) setDocsLoading(false);
        }
    };

    useEffect(() => {
        setGraphData(null);
        const fetchDetails = async () => {
            if (!notebookId) {
                setError("Invalid Notebook ID");
                setLoading(false);
                return;
            }
            try {
                setLoading(true);
                const nb = await api.getNotebook(notebookId);
                setNotebook(nb);
                setError(null);
            } catch (error) {
                console.error(error);
                setError("Failed to load notebook. It might not exist or the server is down.");
            } finally {
                setLoading(false);
            }
        };
        fetchDetails();
        fetchDocuments(false);
        
        // Poll documents
        const interval = setInterval(() => fetchDocuments(true), 5000);
        return () => clearInterval(interval);
    }, [notebookId]);

    useEffect(() => {
        if (documents.length > 0 && documents[0].status === 'completed') {
            // Only fetch if we don't have data or if the data is for a different document (though here we assume 1 doc per notebook)
            // Or if we want to refresh. For now, let's just fetch.
            // To avoid infinite loops if setGraphData triggers something, but it shouldn't.
            // Better: check if we already have graph data for this doc? 
            // For simplicity, we just fetch. The API call is cheap if cached, but here it's a DB call.
            // Let's add a small check to avoid re-fetching if we already have data and the doc ID matches?
            // But graphData doesn't have doc ID easily accessible unless we store it.
            // Let's just fetch for now, but maybe debounce or check if graphData is null.
            if (!graphData) {
                 api.getDocumentGraph(documents[0].id)
                    .then(data => {
                        console.log("Graph data loaded:", data);
                        // Transform data for react-force-graph-2d
                        const formattedData = {
                            nodes: data.nodes || [],
                            links: (data.edges || []).map((edge: any) => ({
                                ...edge,
                                source: edge.source_node_id,
                                target: edge.target_node_id
                            }))
                        };
                        setGraphData(formattedData);
                    })
                    .catch(err => console.error("Failed to load graph:", err));
            }
        } else if (documents.length === 0 || documents[0].status !== 'completed') {
            setGraphData(null);
        }
    }, [documents, graphData]); // Added graphData to dependency to allow the check

    const handleUploadSuccess = () => {
        fetchDocuments(false);
    };

    if (loading) return <div className="p-8 text-center text-gray-400">Loading notebook...</div>;
    if (error) return (
        <div className="p-8 text-center">
            <div className="text-red-400 mb-4">{error}</div>
            <Link to="/" className="text-indigo-400 hover:underline">Return to Dashboard</Link>
        </div>
    );
    if (!notebook) return (
        <div className="p-8 text-center text-gray-400">
            Notebook not found.
            <br />
            <Link to="/" className="text-indigo-400 hover:underline mt-4 inline-block">Return to Dashboard</Link>
        </div>
    );

    return (
        <div className="h-screen flex flex-col bg-gray-900 text-gray-100">
            {/* Header */}
            <header className="bg-gray-800 shadow-sm border-b border-gray-700 px-6 py-4 flex justify-between items-center">
                <div className="flex items-center">
                    <Link to="/" className="text-indigo-400 hover:text-indigo-300 mr-4">
                        &larr; Back
                    </Link>
                    <h1 className="text-xl font-bold text-gray-100">{notebook.title}</h1>
                </div>
            </header>

            {/* Main Content */}
            <div className="flex-1 flex overflow-hidden">
                {/* Left: Chat */}
                <div className="w-1/2 border-r border-gray-700 bg-gray-900 flex flex-col">
                    <div className="flex-1 p-4 overflow-y-auto">
                        {messages.map((msg, idx) => (
                            <div key={idx} className={`mb-4 p-4 rounded-lg shadow-sm border ${msg.role === 'user' ? 'bg-indigo-900/30 border-indigo-800 ml-8' : 'bg-gray-800 border-gray-700 mr-8'}`}>
                                <div className="prose prose-invert max-w-none text-gray-300 text-sm">
                                    <ReactMarkdown>{msg.content}</ReactMarkdown>
                                </div>
                            </div>
                        ))}
                        {isChatting && (
                            <div className="mb-4 p-4 rounded-lg shadow-sm border bg-gray-800 border-gray-700 mr-8 animate-pulse">
                                <div className="h-4 bg-gray-700 rounded w-3/4 mb-2"></div>
                                <div className="h-4 bg-gray-700 rounded w-1/2"></div>
                            </div>
                        )}
                        <div ref={messagesEndRef} />
                    </div>
                    <div className="p-4 border-t border-gray-700 bg-gray-800">
                        <form onSubmit={handleSendMessage}>
                            <div className="flex gap-2">
                                <input
                                    type="text"
                                    value={input}
                                    onChange={(e) => setInput(e.target.value)}
                                    placeholder="Ask a question about your documents..."
                                    className="flex-1 bg-gray-700 border border-gray-600 rounded-md p-2 text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                                    disabled={isChatting}
                                />
                                <button 
                                    type="submit"
                                    disabled={isChatting || !input.trim()}
                                    className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                >
                                    Send
                                </button>
                            </div>
                        </form>
                        <p className="text-xs text-gray-500 mt-2 text-center">AI can make mistakes. Please verify important information.</p>
                    </div>
                </div>

                {/* Right: Documents & Graph */}
                <div className="w-1/2 flex flex-col bg-gray-800 border-l border-gray-700">
                    <div className="p-4 border-b border-gray-700 flex justify-between items-center">
                        <h2 className="font-medium text-gray-100">Documents</h2>
                        {documents.length === 0 && (
                            <UploadForm onUploadSuccess={handleUploadSuccess} notebookId={notebookId} />
                        )}
                    </div>
                    
                    <div className="flex-1 overflow-y-auto p-4 bg-gray-900">
                        <DocumentList documents={documents} loading={docsLoading} error={docsError} />
                    </div>
                    
                    <div className="h-1/2 border-t border-gray-700 bg-gray-900 relative flex flex-col" ref={graphContainerRef}>
                        <div className="absolute top-0 left-0 right-0 p-2 bg-gray-800/80 z-10 flex justify-between items-center border-b border-gray-700 backdrop-blur-sm">
                            <h3 className="text-sm font-bold text-gray-200">Knowledge Graph</h3>
                            {graphData && (
                                <div className="flex gap-2 text-xs">
                                    <span className="bg-blue-900/50 text-blue-200 px-2 py-0.5 rounded border border-blue-800">Nodes: {graphData.nodes?.length || 0}</span>
                                    <span className="bg-green-900/50 text-green-200 px-2 py-0.5 rounded border border-green-800">Edges: {graphData.links?.length || 0}</span>
                                </div>
                            )}
                        </div>
                        
                        {graphData ? (
                            <div className="flex-1 w-full h-full overflow-hidden">
                                <ForceGraph2D
                                    width={graphDimensions.width}
                                    height={graphDimensions.height}
                                    graphData={graphData}
                                    nodeLabel="name"
                                    nodeAutoColorBy="label"
                                    linkDirectionalArrowLength={3.5}
                                    linkDirectionalArrowRelPos={1}
                                    backgroundColor="#111827" // gray-900
                                    linkColor={() => "#4b5563"} // gray-600
                                />
                            </div>
                        ) : (
                            <div className="h-full flex items-center justify-center text-gray-500">
                                {documents.length > 0 && documents[0].status === 'processing' ? (
                                    <div className="flex flex-col items-center animate-pulse">
                                        <span className="mb-2 text-indigo-400">Processing Document...</span>
                                        <span className="text-xs">Extracting Entities & Relations</span>
                                    </div>
                                ) : documents.length > 0 && documents[0].status === 'completed' ? (
                                    'Loading Graph...'
                                ) : (
                                    'Upload a document to generate Knowledge Graph'
                                )}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};
