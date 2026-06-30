package constants

const DBCreationSchema = `
CREATE TABLE IF NOT EXISTS teams (
    team_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    team_name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS tournaments (
    tournament_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS matches (
    match_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tournament_id INT NOT NULL,
    round_number INT NOT NULL, 
    match_position INT NOT NULL, 

    team1_id INT NULL,
    team2_id INT NULL,
    winner_id INT NULL,

    team1_prev_match_id INT NULL,
    team2_prev_match_id INT NULL,

    CONSTRAINT fk_tournament FOREIGN KEY (tournament_id) REFERENCES tournaments(tournament_id) ON DELETE CASCADE,
    CONSTRAINT fk_team1 FOREIGN KEY (team1_id) REFERENCES teams(team_id),
    CONSTRAINT fk_team2 FOREIGN KEY (team2_id) REFERENCES teams(team_id),
    CONSTRAINT fk_winner FOREIGN KEY (winner_id) REFERENCES teams(team_id),
    CONSTRAINT fk_team1_prev_match FOREIGN KEY (team1_prev_match_id) REFERENCES matches(match_id) ON DELETE SET NULL,
    CONSTRAINT fk_team2_prev_match FOREIGN KEY (team2_prev_match_id) REFERENCES matches(match_id) ON DELETE SET NULL,
    CONSTRAINT uniq_match_pos UNIQUE (tournament_id, round_number, match_position)
);

CREATE TABLE IF NOT EXISTS users (
    user_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_brackets (
    user_bracket_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT NOT NULL,
    tournament_id INT NOT NULL,
    is_master BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_user_tournament FOREIGN KEY (tournament_id) REFERENCES tournaments(tournament_id) ON DELETE CASCADE,
    CONSTRAINT uniq_user_tournament UNIQUE (user_id, tournament_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_brackets_master_tournament ON user_brackets(tournament_id) WHERE is_master;

CREATE TABLE IF NOT EXISTS match_predictions (
    prediction_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_bracket_id INT NOT NULL,
    match_id INT NOT NULL,
    predicted_winner_id INT NULL,

    CONSTRAINT fk_user_bracket FOREIGN KEY (user_bracket_id) REFERENCES user_brackets(user_bracket_id) ON DELETE CASCADE,
    CONSTRAINT fk_predicted_match FOREIGN KEY (match_id) REFERENCES matches(match_id) ON DELETE CASCADE,
    CONSTRAINT fk_predicted_winner FOREIGN KEY (predicted_winner_id) REFERENCES teams(team_id),
    CONSTRAINT uniq_bracket_match UNIQUE (user_bracket_id, match_id)
);

CREATE INDEX IF NOT EXISTS idx_matches_tournament ON matches(tournament_id);
CREATE INDEX IF NOT EXISTS idx_match_predictions_bracket ON match_predictions(user_bracket_id);
CREATE INDEX IF NOT EXISTS idx_match_predictions_match ON match_predictions(match_id);
`
