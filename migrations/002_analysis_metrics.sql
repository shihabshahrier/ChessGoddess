-- Migration: 002_analysis_metrics.sql
-- Richer per-move analysis: centipawn loss, win-probability accuracy, engine lines.

ALTER TABLE moves
    ADD COLUMN IF NOT EXISTS eval_before DOUBLE PRECISION DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS eval_after  DOUBLE PRECISION DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS cp_loss     DOUBLE PRECISION DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS accuracy    DOUBLE PRECISION DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS best_line   TEXT DEFAULT '';

-- Widen the classification vocabulary (brilliant / great / book added).
ALTER TABLE moves DROP CONSTRAINT IF EXISTS moves_classification_check;
ALTER TABLE moves ADD CONSTRAINT moves_classification_check
    CHECK (classification IN (
        'brilliant', 'great', 'best', 'excellent', 'good',
        'book', 'inaccuracy', 'mistake', 'blunder'
    ));

-- best_move now stores UCI long-algebraic (e.g. e2e4, e7e8q) — widen it.
ALTER TABLE moves ALTER COLUMN best_move TYPE VARCHAR(12);

ALTER TABLE analysis_sessions
    ADD COLUMN IF NOT EXISTS accuracy_white DOUBLE PRECISION DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS accuracy_black DOUBLE PRECISION DEFAULT 0.0;
