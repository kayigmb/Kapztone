DROP TABLE rule_events;

ALTER TABLE flags DROP COLUMN source, DROP COLUMN assignee_user_id;

DROP TRIGGER handovers_template_version ON handovers;
DROP FUNCTION set_handover_template_version();
DROP TRIGGER transforms_template_version ON transforms;
DROP FUNCTION set_transform_template_version();
ALTER TABLE handovers DROP COLUMN template_version;
ALTER TABLE transforms DROP COLUMN template_version;

ALTER TABLE commodity_templates DROP COLUMN draft, DROP COLUMN version;

ALTER TABLE facilities
    DROP COLUMN approved_at,
    DROP COLUMN approved_by,
    DROP COLUMN licence_storage_key,
    DROP COLUMN licence_sha256,
    DROP COLUMN licence_size_bytes,
    DROP COLUMN licence_content_type,
    DROP COLUMN licence_file_name,
    DROP COLUMN licence_valid_to;
