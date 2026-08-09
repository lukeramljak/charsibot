CREATE TABLE IF NOT EXISTS `user_plushies` (
	`user_id` text NOT NULL,
	`username` text NOT NULL,
	`series` text NOT NULL,
	`key` text NOT NULL,
	PRIMARY KEY(`user_id`, `series`, `key`)
);
--> statement-breakpoint
CREATE TABLE IF NOT EXISTS `user_stats` (
	`user_id` text NOT NULL,
	`username` text NOT NULL,
	`stat_name` text NOT NULL,
	`value` integer DEFAULT 3 NOT NULL,
	PRIMARY KEY(`user_id`, `stat_name`)
);
--> statement-breakpoint
CREATE TABLE IF NOT EXISTS `viewer_activity` (
	`user_id` text PRIMARY KEY NOT NULL,
	`username` text NOT NULL,
	`last_active_at` text NOT NULL
);
--> statement-breakpoint
CREATE INDEX IF NOT EXISTS `viewer_activity_last_active_at_idx` ON `viewer_activity` (`last_active_at`);
