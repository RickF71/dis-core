// Simplified Cedar schema
entity user {}
entity domain {}

action freeze { appliesTo: { principal: user, resource: domain } }
action unfreeze { appliesTo: { principal: user, resource: domain } }

policy allow_freeze when { principal has role == "seat:authority" };
