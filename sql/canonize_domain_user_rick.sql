-- ============================================================
-- CANONIZE DOMAIN.USER.RICK
-- Promotes the validated bootstrap YAML into immutable canon
-- ============================================================

-- Remove any previous record (optional safety)
DELETE FROM canon
WHERE type = 'domain'
  AND content->'meta'->>'domain_id' = 'domain.user.rick';

-- Insert the new canonical version
INSERT INTO canon (type, content)
VALUES (
  'domain',
  '{
    "meta": {
      "name": "Rick Personal Domain",
      "domain_id": "domain.user.rick",
      "schema_id": "domain.person",
      "description": "Rick’s personal sovereign domain root.",
      "schema_version": "v1.0"
    },
    "seats": [{
      "id": "seat.root",
      "type": "root",
      "holder": "rick",
      "authority": "authoritant",
      "description": "Root seat and primary moral agent for this domain.",
      "permissions": [
        "manage_visuals",
        "manage_subdomains",
        "manage_identity"
      ]
    }],
    "parents": ["domain.usa", "domain.terra"],
    "settings": {
      "visual": {
        "border_color": "#FFD60A",
        "border_style": "solid",
        "border_width": "8px"
      },
      "permissions": { "allow_self_override": true }
    },
    "interface": {
      "dis_css": {
        "ref": "domain.terra",
        "append": "body { border-top: 8px solid limegreen !important; background-color: #f0fff0; }"
      },
      "dis_jsx": {
        "ref": "domain.terra",
        "append": "<div>Welcome Rick</div>"
      }
    }
  }'::jsonb
);

