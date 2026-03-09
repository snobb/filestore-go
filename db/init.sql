
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE doc_status AS ENUM ('pending', 'uploaded', 'verified', 'rejected');

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    file_name TEXT UNIQUE NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    content_type TEXT NOT NULL,
    checksum TEXT,
    status doc_status DEFAULT 'pending',
    uploaded_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- optimistic idempotency control.
ALTER TABLE documents ADD CONSTRAINT unique_user_filename
UNIQUE (user_id, file_name);

ALTER TABLE documents ADD CONSTRAINT fk_documents_user
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Demo user
INSERT INTO users (id, email, username, hashed_password)
VALUES ('00000000-0000-0000-0000-000000000000', 'test@snobb.org', 'DemoClient', 'nopassword');
