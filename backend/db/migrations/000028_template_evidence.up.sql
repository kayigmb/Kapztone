-- Templates drop their stages: a template keeps only the evidence the origin
-- facility must hold before issuing, taken from its ISSUE stage. Evidence
-- later stages asked for was no longer checked and is dropped.
ALTER TABLE commodity_templates ADD COLUMN required_evidence text[] NOT NULL DEFAULT '{}';

UPDATE commodity_templates t SET required_evidence = ARRAY(
    SELECT DISTINCT ev FROM jsonb_array_elements(t.stages) st,
        jsonb_array_elements_text(COALESCE(st->'requiredEvidence', '[]'::jsonb)) ev
    WHERE st->'events' ? 'ISSUE' ORDER BY ev);

UPDATE commodity_templates SET draft = (draft - 'stages') || jsonb_build_object('requiredEvidence', ARRAY(
    SELECT DISTINCT ev FROM jsonb_array_elements(COALESCE(draft->'stages', '[]'::jsonb)) st,
        jsonb_array_elements_text(COALESCE(st->'requiredEvidence', '[]'::jsonb)) ev
    WHERE st->'events' ? 'ISSUE' ORDER BY ev))
WHERE draft IS NOT NULL;

ALTER TABLE commodity_templates DROP COLUMN stages;
