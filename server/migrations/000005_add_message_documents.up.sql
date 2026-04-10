CREATE TABLE IF NOT EXISTS message_documents (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    PRIMARY KEY (message_id, document_id)
);

CREATE INDEX IF NOT EXISTS idx_message_documents_document_id ON message_documents(document_id);
