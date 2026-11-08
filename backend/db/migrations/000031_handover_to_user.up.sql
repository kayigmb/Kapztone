-- The person a handover is sent to (an admin or member of the receiving
-- organisation). Older handovers have none: anyone there could answer.
ALTER TABLE handovers ADD COLUMN to_user uuid REFERENCES users (id);
