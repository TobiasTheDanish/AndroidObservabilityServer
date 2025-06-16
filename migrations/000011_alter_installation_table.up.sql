BEGIN;
ALTER TABLE ob_installations
	 ADD COLUMN type text NOT NULL DEFAULT 'android',
	 ADD COLUMN data JSONB;

UPDATE ob_installations AS i1
	SET data = (
		SELECT json_build_object(
			'brand', i2.brand,
			'model', i2.model,
			'sdkVersion', i2.sdk_version
		) FROM ob_installations AS i2
		WHERE i1.id = i2.id
	);

ALTER TABLE ob_installations
	 DROP COLUMN brand,
	 DROP COLUMN model,
	 DROP COLUMN sdk_version;

COMMIT;
