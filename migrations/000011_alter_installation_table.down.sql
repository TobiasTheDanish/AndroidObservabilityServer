BEGIN;

ALTER TABLE ob_installations
	 ADD COLUMN sdk_version INTEGER DEFAULT -1,
	 ADD COLUMN model TEXT DEFAULT '',
   ADD COLUMN brand TEXT DEFAULT '';

UPDATE ob_installations AS i1
	SET sdk_version = (
		SELECT CAST ((data ->> 'sdkVersion') AS INTEGER) FROM ob_installations AS i2
		WHERE i1.id = i2.id
	), brand = (
		SELECT data ->> 'brand' FROM ob_installations AS i2
		WHERE i1.id = i2.id
	), model = (
		SELECT data ->> 'model' FROM ob_installations AS i2
		WHERE i1.id = i2.id
	);

ALTER TABLE ob_installations
	 DROP COLUMN type,
	 DROP COLUMN data;

COMMIT;
