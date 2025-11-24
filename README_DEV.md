# Development README

## Reset + Bootstrap (Atomic Invite-Accept Test)

Follow these steps to reset the local database, restart `dis-core`, and exercise the new atomic invite-accept flow.

1. Reset the database (this will drop and recreate the `dis` database):

   ./scripts/reset_db.sh

2. Restart dis-core (set `DIS_DB_PASS` in the environment first):

   export DIS_DB_PASS="your_db_password"
   ./scripts/restart_dis_core.sh

   Note: the restart script builds `dis-core` then runs the binary.

3. Generate a handshake invite (example):

   curl -X POST http://localhost:8080/api/invite/new \
     -H "Content-Type: application/json" \
     -d '{"subject":"rick"}'

   The response will include a token (or the server logs will show it). Save the token for the next step.

4. Accept the invite (replace <TOKEN_FROM_PREVIOUS> with the invite token):

   curl -X POST http://localhost:8080/api/invite/accept \
     -H "Content-Type: application/json" \
     -d '{"token":"<TOKEN_FROM_PREVIOUS>", "presentation_name":"Rick"}'

The invite accept response should now include `identity_id`, `domain_id`, `actor_id`, and `receipt_id`.
