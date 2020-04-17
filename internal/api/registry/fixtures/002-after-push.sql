INSERT INTO accounts (name, auth_tenant_id, upstream_peer_hostname, required_labels, blobs_sweeped_at, storage_sweeped_at, metadata_json, announced_to_federation_at) VALUES ('test1', 'test1authtenant', '', '', NULL, NULL, '', NULL);

INSERT INTO blob_mounts (blob_id, repo_id, marked_for_deletion_at) VALUES (1, 1, NULL);

INSERT INTO blobs (id, account_name, digest, size_bytes, storage_id, pushed_at, validated_at, validation_error_message, marked_for_deletion_at) VALUES (1, 'test1', 'sha256:7575de20fdeacfb9a529c26f03c5beab401bb985721b923bba6b34fe4fce5f9c', 1486, '6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b', 3, 3, '', NULL);

INSERT INTO manifest_blob_refs (repo_id, digest, blob_id) VALUES (1, 'sha256:86fa8722ca7f27e97e1bc5060c3f6720bf43840f143f813fcbe48ed4cbeebb90', 1);

INSERT INTO manifests (repo_id, digest, media_type, size_bytes, pushed_at, validated_at, validation_error_message) VALUES (1, 'sha256:86fa8722ca7f27e97e1bc5060c3f6720bf43840f143f813fcbe48ed4cbeebb90', 'application/vnd.docker.distribution.manifest.v2+json', 1751, 3, 3, '');

INSERT INTO quotas (auth_tenant_id, manifests) VALUES ('test1authtenant', 100);

INSERT INTO repos (id, account_name, name, blob_mounts_sweeped_at, manifests_synced_at) VALUES (1, 'test1', 'foo', NULL, NULL);

INSERT INTO tags (repo_id, name, digest, pushed_at) VALUES (1, 'latest', 'sha256:86fa8722ca7f27e97e1bc5060c3f6720bf43840f143f813fcbe48ed4cbeebb90', 3);
