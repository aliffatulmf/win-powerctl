/**
 * poweroff.c - Windows Shutdown Library
 *
 * Provides programmatic control over system power states (shutdown, reboot,
 * power off) with internal privilege escalation and dry-run validation.
 *
 * Compilation:
 *   MSVC  : cl /LD poweroff.c /link /OUT:poweroff.dll advapi32.lib user32.lib
 *   MinGW : gcc -shared -o poweroff.dll poweroff.c -ladvapi32 -luser32
 */

#ifndef POWEROFF_C
#define POWEROFF_C

#include <windows.h>

/**
 * Enables the SeShutdownPrivilege for the current process token.
 *
 * @return TRUE if the privilege was successfully enabled, FALSE otherwise.
 *         Call GetLastError() for extended error information.
 */
static BOOL enable_shutdown_privilege(void)
{
    HANDLE token;
    LUID luid;
    TOKEN_PRIVILEGES tp;
    BOOL ok;
    DWORD err;

    if (!OpenProcessToken(GetCurrentProcess(), TOKEN_ADJUST_PRIVILEGES | TOKEN_QUERY, &token))
    {
        return FALSE;
    }

    /* Use hardcode wide-string to ensure compatibility across all Windows SDK versions */
    if (!LookupPrivilegeValueW(NULL, L"SeShutdownPrivilege", &luid))
    {
        CloseHandle(token);
        return FALSE;
    }

    tp.PrivilegeCount = 1;
    tp.Privileges[0].Luid = luid;
    tp.Privileges[0].Attributes = SE_PRIVILEGE_ENABLED;

    ok = AdjustTokenPrivileges(token, FALSE, &tp, 0, NULL, NULL);
    err = GetLastError();
    CloseHandle(token);

    /* AdjustTokenPrivileges returns TRUE even if not all privileges are assigned. */
    return ok && err != ERROR_NOT_ALL_ASSIGNED;
}

/**
 * Internal worker to handle privilege acquisition and system shutdown execution.
 *
 * @param flags    The ExitWindowsEx action flags (e.g., EWX_SHUTDOWN, EWX_REBOOT).
 * @param dry_run  If TRUE, validates privileges without initiating the shutdown.
 * @return TRUE on success or successful dry-run, FALSE on failure.
 */
static BOOL do_shutdown(UINT flags, BOOL dry_run)
{
    if (!enable_shutdown_privilege())
    {
        return FALSE;
    }

    if (dry_run)
    {
        return TRUE;
    }

    return ExitWindowsEx(flags, SHTDN_REASON_FLAG_PLANNED);
}

#ifdef __cplusplus
extern "C"
{
#endif

    /**
     * Enables the shutdown privilege for the calling process.
     *
     * @return TRUE on success, FALSE on failure.
     */
    __declspec(dllexport) BOOL WINAPI EnableShutdownPrivilege(void)
    {
        return enable_shutdown_privilege();
    }

    /**
     * Initiates a graceful system shutdown.
     *
     * @param dry_run  If TRUE, performs privilege validation only.
     * @return TRUE on success, FALSE on failure.
     */
    __declspec(dllexport) BOOL WINAPI ShutdownGraceful(BOOL dry_run)
    {
        return do_shutdown(EWX_SHUTDOWN, dry_run);
    }

    /**
     * Initiates a forced system shutdown, terminating unresponsive applications.
     *
     * @param dry_run  If TRUE, performs privilege validation only.
     * @return TRUE on success, FALSE on failure.
     */
    __declspec(dllexport) BOOL WINAPI ShutdownForce(BOOL dry_run)
    {
        return do_shutdown(EWX_SHUTDOWN | EWX_FORCE | EWX_FORCEIFHUNG, dry_run);
    }

    /**
     * Initiates a system reboot.
     *
     * @param dry_run  If TRUE, performs privilege validation only.
     * @return TRUE on success, FALSE on failure.
     */
    __declspec(dllexport) BOOL WINAPI ShutdownReboot(BOOL dry_run)
    {
        return do_shutdown(EWX_REBOOT | EWX_FORCEIFHUNG, dry_run);
    }

    /**
     * Initiates a system power off.
     *
     * @param dry_run  If TRUE, performs privilege validation only.
     * @return TRUE on success, FALSE on failure.
     */
    __declspec(dllexport) BOOL WINAPI ShutdownPowerOff(BOOL dry_run)
    {
        return do_shutdown(EWX_POWEROFF | EWX_FORCEIFHUNG, dry_run);
    }

#ifdef __cplusplus
}
#endif

#endif /* POWEROFF_C */
