-- Back to one origin stage that issues and asks for the template's evidence.
ALTER TABLE commodity_templates ADD COLUMN stages jsonb;

UPDATE commodity_templates SET stages = jsonb_build_array(jsonb_build_object(
    'key', 'origin', 'events', jsonb_build_array('ISSUE'), 'requiredEvidence', to_jsonb(required_evidence)));

UPDATE commodity_templates SET draft = (draft - 'requiredEvidence') || jsonb_build_object('stages', jsonb_build_array(jsonb_build_object(
    'key', 'origin', 'events', jsonb_build_array('ISSUE'), 'requiredEvidence', COALESCE(draft->'requiredEvidence', '[]'::jsonb))))
WHERE draft IS NOT NULL;

ALTER TABLE commodity_templates ALTER COLUMN stages SET NOT NULL, DROP COLUMN required_evidence;
