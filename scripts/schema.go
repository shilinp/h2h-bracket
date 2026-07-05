package main

const DBCreationSchema = `
CREATE TABLE IF NOT EXISTS teams (
    team_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    team_name VARCHAR(100) NOT NULL UNIQUE,
    is_bye BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    user_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS global_settings (
    is_locked BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS matches (
    match_id INT PRIMARY KEY,
    round_number INT NOT NULL, 
    visual_position INT NOT NULL, 

    team1_id INT NULL,
    team2_id INT NULL,

    team1_prev_match_id INT NULL,
    team2_prev_match_id INT NULL,

    CONSTRAINT fk_team1 FOREIGN KEY (team1_id) REFERENCES teams(team_id),
    CONSTRAINT fk_team2 FOREIGN KEY (team2_id) REFERENCES teams(team_id),
    CONSTRAINT fk_team1_prev_match FOREIGN KEY (team1_prev_match_id) REFERENCES matches(match_id) ON DELETE SET NULL,
    CONSTRAINT fk_team2_prev_match FOREIGN KEY (team2_prev_match_id) REFERENCES matches(match_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS match_predictions (
    user_id INT NOT NULL,
    match_id INT NOT NULL,
    predicted_winner_id INT NULL,

    PRIMARY KEY (user_id, match_id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_match FOREIGN KEY (match_id) REFERENCES matches(match_id) ON DELETE CASCADE,
    CONSTRAINT fk_winner FOREIGN KEY (predicted_winner_id) REFERENCES teams(team_id)
);

CREATE INDEX IF NOT EXISTS idx_match_predictions_user ON match_predictions(user_id);
`
