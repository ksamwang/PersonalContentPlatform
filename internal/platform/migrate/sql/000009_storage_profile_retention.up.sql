ALTER TABLE upload_intents DROP CONSTRAINT IF EXISTS upload_intents_storage_profile_id_fkey;
ALTER TABLE upload_intents ADD CONSTRAINT upload_intents_storage_profile_id_fkey FOREIGN KEY(storage_profile_id) REFERENCES storage_profiles(id) ON DELETE SET NULL;

ALTER TABLE blobs DROP CONSTRAINT IF EXISTS blobs_storage_profile_id_fkey;
ALTER TABLE blobs ADD CONSTRAINT blobs_storage_profile_id_fkey FOREIGN KEY(storage_profile_id) REFERENCES storage_profiles(id) ON DELETE SET NULL;
