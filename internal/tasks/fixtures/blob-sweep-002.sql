INSERT INTO accounts (name, auth_tenant_id, upstream_peer_hostname, required_labels, metadata_json, next_blob_sweep_at, next_storage_sweep_at, next_federation_announcement_at, in_maintenance, external_peer_url, external_peer_username, external_peer_password) VALUES ('test1', 'test1authtenant', '', '', '', 14400, NULL, NULL, FALSE, '', '', '');

INSERT INTO blob_mounts (blob_id, repo_id, can_be_deleted_at) VALUES (4, 1, NULL);
INSERT INTO blob_mounts (blob_id, repo_id, can_be_deleted_at) VALUES (5, 1, NULL);

INSERT INTO blobs (id, account_name, digest, size_bytes, storage_id, pushed_at, validated_at, validation_error_message, can_be_deleted_at) VALUES (1, 'test1', 'sha256:3d309a098968afe810ee167e0c5b205ef3610829a6d34e0f0ba4ca66756c6f5e', 1048576, '3d309a098968afe810ee167e0c5b205ef3610829a6d34e0f0ba4ca66756c6f5e', 3600, 3600, '', 12600);
INSERT INTO blobs (id, account_name, digest, size_bytes, storage_id, pushed_at, validated_at, validation_error_message, can_be_deleted_at) VALUES (2, 'test1', 'sha256:a718f4a112454b50c8ecd2b0a5b00eb32ee90699593625139cd3fabc97dcce8d', 1048576, 'a718f4a112454b50c8ecd2b0a5b00eb32ee90699593625139cd3fabc97dcce8d', 3600, 3600, '', 12600);
INSERT INTO blobs (id, account_name, digest, size_bytes, storage_id, pushed_at, validated_at, validation_error_message, can_be_deleted_at) VALUES (3, 'test1', 'sha256:d87b0830e55e19fd0825bebaa110ebade6de7d1fcf2ddf0fbd5762e1f809370e', 1048576, 'd87b0830e55e19fd0825bebaa110ebade6de7d1fcf2ddf0fbd5762e1f809370e', 3600, 3600, '', 12600);
INSERT INTO blobs (id, account_name, digest, size_bytes, storage_id, pushed_at, validated_at, validation_error_message, can_be_deleted_at) VALUES (4, 'test1', 'sha256:4ebfd1aafe77056e348c3ef48fa229b60390abb9c15c1023f609d88c06943eab', 1048576, '4ebfd1aafe77056e348c3ef48fa229b60390abb9c15c1023f609d88c06943eab', 3600, 3600, '', NULL);
INSERT INTO blobs (id, account_name, digest, size_bytes, storage_id, pushed_at, validated_at, validation_error_message, can_be_deleted_at) VALUES (5, 'test1', 'sha256:65c59193f05fcb20c54ad7ac25585907a04553033a9a61dbd32ced96fdd24fe1', 1048576, '65c59193f05fcb20c54ad7ac25585907a04553033a9a61dbd32ced96fdd24fe1', 3600, 3600, '', NULL);

INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (1, 'test1', 'foo', NULL, NULL);
