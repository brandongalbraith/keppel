INSERT INTO accounts (name, auth_tenant_id, upstream_peer_hostname, required_labels, metadata_json, next_blob_sweep_at, next_storage_sweep_at, next_federation_announcement_at, in_maintenance, external_peer_url, external_peer_username, external_peer_password) VALUES ('test1', 'tenant1', '', '', '', NULL, NULL, NULL, FALSE, '', '', '');
INSERT INTO accounts (name, auth_tenant_id, upstream_peer_hostname, required_labels, metadata_json, next_blob_sweep_at, next_storage_sweep_at, next_federation_announcement_at, in_maintenance, external_peer_url, external_peer_username, external_peer_password) VALUES ('test2', 'tenant2', '', '', '', NULL, NULL, NULL, FALSE, '', '', '');

INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:040a5a009f9b9d5e4771742174142e74fa2d3e0aaa3df5717f01ade338d75d0e', '', 9000, 10090, 10090, '', NULL);
INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:04abc8821a06e5a30937967d11ad10221cb5ac3b5273e434f1284ee87129a061', '', 8000, 10080, 10080, '', NULL);
INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:27ecd0a598e76f8a2fd264d427df0a119903e8eae384e478902541756f089dd1', '', 4000, 10040, 10040, '', NULL);
INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:377a23f52c6b357696238c3318f677a082dd3430bb6691042bd550a5cda28ebb', '', 5000, 10050, 10050, '', NULL);
INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:4bf5122f344554c53bde2ebb8cd2b7e3d1600ad631c385a5d7cce23c7785459a', '', 1000, 10010, 10010, '', NULL);
INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:75c8fd04ad916aec3e3d5cb76a452b116b3d4d0912a0a485e9fb8e3d240e210c', '', 3000, 10030, 10030, '', NULL);
INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:9dcf97a184f32623d11a73124ceb99a5709b083721e878a16d78f596718ba7b2', '', 2000, 10020, 10020, '', NULL);
INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:c36336f242c655c52fa06c4d03f665ca9ea0bb84f20f1b1f90976aa58ca40a4a', '', 7000, 10070, 10070, '', NULL);
INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:dfea2964b5deedea7b1ef077de529c3959e6788bdbb3441e70c77a1ae875bb48', '', 6000, 10060, 10060, '', NULL);
INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message, last_pulled_at) VALUES (5, 'sha256:ffadf8d89d37b3b55fe1847b513cf92e3be87e4c168708c7851845df96fb36be', '', 10000, 10100, 10100, '', NULL);

INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (1, 'test1', 'repo1-1', NULL, NULL);
INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (10, 'test2', 'repo2-5', NULL, NULL);
INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (2, 'test2', 'repo2-1', NULL, NULL);
INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (3, 'test1', 'repo1-2', NULL, NULL);
INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (4, 'test2', 'repo2-2', NULL, NULL);
INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (5, 'test1', 'repo1-3', NULL, NULL);
INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (6, 'test2', 'repo2-3', NULL, NULL);
INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (7, 'test1', 'repo1-4', NULL, NULL);
INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (8, 'test2', 'repo2-4', NULL, NULL);
INSERT INTO repos (id, account_name, name, next_blob_mount_sweep_at, next_manifest_sync_at) VALUES (9, 'test1', 'repo1-5', NULL, NULL);

INSERT INTO tags (repo_id, name, digest, pushed_at, last_pulled_at) VALUES (5, 'tag1', 'sha256:4bf5122f344554c53bde2ebb8cd2b7e3d1600ad631c385a5d7cce23c7785459a', 20010, NULL);
INSERT INTO tags (repo_id, name, digest, pushed_at, last_pulled_at) VALUES (5, 'tag2', 'sha256:9dcf97a184f32623d11a73124ceb99a5709b083721e878a16d78f596718ba7b2', 20020, NULL);
INSERT INTO tags (repo_id, name, digest, pushed_at, last_pulled_at) VALUES (5, 'tag3', 'sha256:75c8fd04ad916aec3e3d5cb76a452b116b3d4d0912a0a485e9fb8e3d240e210c', 20030, NULL);
