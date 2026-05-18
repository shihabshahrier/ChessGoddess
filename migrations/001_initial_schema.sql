-- Migration: 001_initial_schema.sql
-- Create core tables for ChessLens

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    google_id VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Games table
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pgn TEXT NOT NULL,
    fen VARCHAR(255),
    white_player VARCHAR(255),
    black_player VARCHAR(255),
    result VARCHAR(10),
    opening VARCHAR(255),
    time_control VARCHAR(50),
    event VARCHAR(255),
    date VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_games_user_id ON games(user_id);

-- Analysis sessions table
CREATE TABLE IF NOT EXISTS analysis_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    engine_config JSONB DEFAULT '{}',
    depth INTEGER DEFAULT 20,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_analysis_sessions_game_id ON analysis_sessions(game_id);
CREATE INDEX idx_analysis_sessions_user_id ON analysis_sessions(user_id);
CREATE INDEX idx_analysis_sessions_status ON analysis_sessions(status);

-- Moves table
CREATE TABLE IF NOT EXISTS moves (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    move_number INTEGER NOT NULL,
    fen VARCHAR(255) NOT NULL,
    san VARCHAR(10) NOT NULL,
    evaluation DOUBLE PRECISION DEFAULT 0.0,
    best_move VARCHAR(10),
    classification VARCHAR(20) CHECK (classification IN ('blunder', 'mistake', 'inaccuracy', 'good', 'excellent', 'best')),
    depth INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_moves_session_id ON moves(session_id);
CREATE INDEX idx_moves_classification ON moves(classification);

-- Snapshots table (immutable)
CREATE TABLE IF NOT EXISTS snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    share_token VARCHAR(64) UNIQUE NOT NULL,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_snapshots_session_id ON snapshots(session_id);
CREATE INDEX idx_snapshots_share_token ON snapshots(share_token);

-- AI explanations table
CREATE TABLE IF NOT EXISTS ai_explanations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    move_id UUID REFERENCES moves(id) ON DELETE SET NULL,
    fen VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    model VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ai_explanations_session_id ON ai_explanations(session_id);
CREATE INDEX idx_ai_explanations_fen ON ai_explanations(fen);

-- Uploads table (R2 references)
CREATE TABLE IF NOT EXISTS uploads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    r2_key VARCHAR(500) NOT NULL,
    url TEXT NOT NULL,
    type VARCHAR(50) CHECK (type IN ('screenshot', 'board_image', 'thumbnail')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_uploads_user_id ON uploads(user_id);
CREATE INDEX idx_uploads_type ON uploads(type);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_games_updated_at BEFORE UPDATE ON games FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_analysis_sessions_updated_at BEFORE UPDATE ON analysis_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
