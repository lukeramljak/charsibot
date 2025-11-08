CREATE TABLE IF NOT EXISTS `stats` (
	`id` text PRIMARY KEY NOT NULL,
	`username` text NOT NULL,
	`strength` integer DEFAULT 3 NOT NULL,
	`intelligence` integer DEFAULT 3 NOT NULL,
	`charisma` integer DEFAULT 3 NOT NULL,
	`luck` integer DEFAULT 3 NOT NULL,
	`dexterity` integer DEFAULT 3 NOT NULL,
	`penis` integer DEFAULT 3 NOT NULL
);
--> statement-breakpoint
CREATE TABLE IF NOT EXISTS `oauth_tokens` (
	`token_type` text PRIMARY KEY NOT NULL,
	`access_token` text NOT NULL,
	`refresh_token` text NOT NULL,
	`updated_at` numeric DEFAULT CURRENT_TIMESTAMP
);
--> statement-breakpoint
CREATE TABLE IF NOT EXISTS `user_collections` (
	`user_id` text,
	`username` text NOT NULL,
	`collection_type` text,
	`reward1` integer DEFAULT 0,
	`reward2` integer DEFAULT 0,
	`reward3` integer DEFAULT 0,
	`reward4` integer DEFAULT 0,
	`reward5` integer DEFAULT 0,
	`reward6` integer DEFAULT 0,
	`reward7` integer DEFAULT 0,
	`reward8` integer DEFAULT 0,
	PRIMARY KEY(`user_id`, `collection_type`)
);
