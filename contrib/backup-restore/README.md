# Harbor Backup and Restore Scripts (Contrib)

**Warning:** These scripts are provided as-is in the `contrib/backup-restore` directory. They are not officially maintained or supported by the Harbor project. Use them at your own risk and ensure you understand their functionality before running them in a production environment.

These scripts (`harbor-backup` and `harbor-restore`) are provided as a convenience for backing up and restoring your Harbor instance. They aim to back up the following components:

* Harbor Database (PostgreSQL)
* Container Registry Data
* Chart Museum Data (if enabled)
* Redis Data (if enabled)
* Secret Keys
* Harbor Configuration (`harbor.yml`)

### Features
Compared to the scripts the harbor project used to have in their repo this set of scripts is more robust in its error handling and also offers features
for not packing the backup into a tarball. This makes it easy to rsync the whole backup directory to a secondary/standby node and restore there.

rsync is used extensively by the script. by leaving the files in the backup directory between runs the downtime for backup is greatly reduced at the
expense of disk space usage.

Supports logging of status messages directly to syslog

## Prerequisites

* **Docker:** These scripts rely on the `docker` command-line interface to interact with Harbor's containers. Ensure Docker is installed and accessible on the machine where you run these scripts.
* **Sufficient Permissions:** You'll need appropriate permissions (e.g., `sudo` or being in the `docker` group) to run Docker commands and perform file system operations.
* **Stopped Harbor Instance:** You must stop your Harbor instance completely before running the `harbor-backup` or `harbor-restore` script to avoid data inconsistencies.

## Usage

### Backup (`harbor-backup`)

1.  **Download the Scripts:** Place the `harbor-backup` script in a location accessible from your Harbor instance. Within the Harbor repository, this would typically be under `contrib/backup-restore/`.

2.  **Make it Executable:**
    ```bash
    chmod +x harbor-backup
    ```

3.  **Run the Backup Script:**
    ```bash
    ./harbor-backup [OPTIONS]
    ```
3.  **Stop Harbor:** Ensure your Harbor instance is completely stopped before proceeding with the backup.

4.  **Options:**
    * `--docker-cmd <command>`: Specify the Docker command to use (default: `docker`).
    * `--db-image <image>`: Specify the Harbor database image to use for the temporary backup container (default: auto-detected). It's generally recommended to let it auto-detect.
    * `--db-path <path>`: Harbor DB data path (default: `/data/database`). Adjust if your deployment uses a different path.
    * `--registry-path <path>`: Registry data path (default: `/data/registry`). Adjust if your deployment uses a different path.
    * `--chart-museum-path <path>`: Chart Museum data path (default: `/data/chart_storage`). Adjust if your deployment uses a different path.
    * `--redis-path <path>`: Redis data path (default: `/data/redis`). Adjust if your deployment uses a different path.
    * `--secret-path <path>`: Secret data path (default: `/data/secret`). Adjust if your deployment uses a different path.
    * `--config-path <path>`: Harbor configuration file path (default: `/etc/harbor/harbor.yml`). Adjust if your deployment uses a different path.
    * `--backup-dir <path>`: Directory where the backup will be stored (default: `harbor_backup`).
    * `--no-archive`: Do not create a `tar.gz` archive of the backup directory. The backup will remain as a directory structure in `$BACKUP_DIR/harbor`.
    * `--use-syslog`: Use syslog for logging output.
    * `--log-level <level>`: Set the logging level (default: `INFO`, options: `DEBUG`, `INFO`, `NOTICE`, `WARNING`, `ERROR`, `CRITICAL`, `ALERT`, `EMERGENCY`).
    * `--help`: Display this help message.

5.  **Backup Location:** By default, the backup will be created in a directory named `harbor_backup` in the current working directory. If the `--no-archive` option is not used, the final backup will be a compressed tarball named `harbor_backup.tar.gz` within the `harbor_backup` directory.

### Restore (`harbor-restore`)

1.  **Download the Scripts:** Place the `harbor-restore` script in a location accessible from your Harbor instance. Within the Harbor repository, this would typically be under `contrib/backup-restore/`.

2.  **Make it Executable:**
    ```bash
    chmod +x harbor-restore
    ```

3.  **Stop Harbor:** Ensure your Harbor instance is completely stopped before proceeding with the restore.

4.  **Run the Restore Script:**
    ```bash
    ./harbor-restore [OPTIONS]
    ```

5.  **Options:** The restore script accepts similar options to the backup script, allowing you to specify the Docker command, database image, data paths, and the backup directory.

    * `--backup-dir <path>`: **Crucially**, this should point to the directory containing your Harbor backup (either the `harbor` subdirectory extracted from the tarball or the `harbor_backup` directory if `--no-archive` was used).
    * `--no-archive`: Use this option if your backup is already extracted into the `$BACKUP_DIR/harbor` directory. If your backup is a `tar.gz` file, **do not** use this option; the script will attempt to extract it.

    *(Other options like `--docker-cmd`, `--db-image`, `--db-path`, `--registry-path`, `--chart-museum-path`, `--redis-path`, `--secret-path`, `--config-path`, `--use-syslog`, and `--log-level` function similarly to the backup script.)*

6.  **Restore Process:** The script will:
    * Start a temporary database container.
    * Extract the backup archive (if not using `--no-archive`).
    * Drop and recreate existing Harbor databases.
    * Restore the database content from the backed-up SQL files.
    * Synchronize the registry, chart museum, Redis, and secret data directories.
    * Restore the Harbor configuration file.
    * Clean up the temporary database container.

7.  **Restart Harbor:** Once the restore script completes successfully, you can restart your Harbor instance.

## Important Notes

* **Backup Consistency:** For a consistent backup, it's recommended to stop your Harbor instance or at least ensure minimal write activity during the backup process.
* **Database Image Tag:** In production environments, it's advisable to use a specific tag for the `--db-image` option in both the backup and restore scripts to ensure consistency.
* **Custom Deployments:** If you have a highly customized Harbor deployment with data stored in non-default locations, you **must** use the appropriate command-line options to point the scripts to the correct paths.
* **Testing:** Always test the backup and restore process in a non-production environment before relying on it for critical data.
* **Unsupported:** Remember that these scripts are provided in the `contrib/backup-restore/` directory. They may not be actively maintained, and you might encounter issues. Contributions and improvements from the community are welcome.

## Contributing

If you find issues or have improvements to these scripts, feel free to submit pull requests to the Harbor project in the `contrib/backup-restore/` directory.
