/**
 * poweroff.c - Windows shutdown library
 *
 * Shared library providing shutdown/restart functions for Windows.
 * Used by win-powerctl via syscall DLL loading.
 *
 * Build (MSVC):
 *   cl /LD poweroff.c /link /OUT:poweroff.dll advapi32.lib user32.lib
 *
 * Build (MinGW):
 *   gcc -shared -o poweroff.dll poweroff.c -ladvapi32 -luser32
 */

#ifndef POWEROFF_C
#define POWEROFF_C

#include <windows.h>

/* ----- Shutdown flags ----- */

#define EXIT_WINDOWS_GRACEFUL  0x00000001
#define EXIT_WINDOWS_REBOOT    0x00000002
#define EXIT_WINDOWS_FORCE     0x00000004
#define EXIT_WINDOWS_POWEROFF  0x00000008
#define EXIT_WINDOWS_FORCEHUNG 0x00000010

#define SHUTDOWN_REASON_PLANNED 0x80000000

/* ----- Internal helpers ----- */

/**
 * enable_shutdown_privilege - Enable SeShutdownPrivilege for current process.
 * Returns TRUE on success, FALSE on failure.
 */
static BOOL enable_shutdown_privilege(void) {
    HANDLE token;
    if (!OpenProcessToken(GetCurrentProcess(),
        TOKEN_ADJUST_PRIVILEGES | TOKEN_QUERY, &token))
        return FALSE;

    LUID luid;
    if (!LookupPrivilegeValueW(NULL, L"SeShutdownPrivilege", &luid)) {
        CloseHandle(token);
        return FALSE;
    }

    TOKEN_PRIVILEGES tp;
    tp.PrivilegeCount = 1;
    tp.Privileges[0].Luid = luid;
    tp.Privileges[0].Attributes = SE_PRIVILEGE_ENABLED;

    BOOL ok = AdjustTokenPrivileges(token, FALSE, &tp, 0, NULL, NULL);
    DWORD err = GetLastError();
    CloseHandle(token);
    return ok && err != ERROR_NOT_ALL_ASSIGNED;
}

/* ----- Exported functions ----- */

/**
 * EnableShutdownPrivilege - Exported wrapper for privilege enablement.
 * Returns TRUE on success, FALSE on failure.
 */
__declspec(dllexport) BOOL EnableShutdownPrivilege(void) {
    return enable_shutdown_privilege();
}

/**
 * ShutdownGraceful - Initiate graceful shutdown.
 *
 * @param dry_run  If non-zero, performs all checks but does not
 *                 actually call ExitWindowsEx. Useful for testing.
 * @return TRUE on success (or dry-run), FALSE on failure.
 */
__declspec(dllexport) BOOL ShutdownGraceful(BOOL dry_run) {
    if (dry_run)
        return TRUE;
    if (!enable_shutdown_privilege())
        return FALSE;
    return ExitWindowsEx(EXIT_WINDOWS_GRACEFUL, SHUTDOWN_REASON_PLANNED);
}

/**
 * ShutdownForce - Initiate forced shutdown (kills hung apps).
 *
 * @param dry_run  If non-zero, performs all checks but does not
 *                 actually call ExitWindowsEx. Useful for testing.
 * @return TRUE on success (or dry-run), FALSE on failure.
 */
__declspec(dllexport) BOOL ShutdownForce(BOOL dry_run) {
    if (dry_run)
        return TRUE;
    if (!enable_shutdown_privilege())
        return FALSE;
    return ExitWindowsEx(
        EXIT_WINDOWS_GRACEFUL | EXIT_WINDOWS_FORCE | EXIT_WINDOWS_FORCEHUNG,
        SHUTDOWN_REASON_PLANNED);
}

/**
 * ShutdownReboot - Initiate reboot.
 *
 * @param dry_run  If non-zero, performs all checks but does not
 *                 actually call ExitWindowsEx. Useful for testing.
 * @return TRUE on success (or dry-run), FALSE on failure.
 */
__declspec(dllexport) BOOL ShutdownReboot(BOOL dry_run) {
    if (dry_run)
        return TRUE;
    if (!enable_shutdown_privilege())
        return FALSE;
    return ExitWindowsEx(EXIT_WINDOWS_REBOOT | EXIT_WINDOWS_FORCEHUNG, SHUTDOWN_REASON_PLANNED);
}

/**
 * ShutdownPowerOff - Initiate power off.
 *
 * @param dry_run  If non-zero, performs all checks but does not
 *                 actually call ExitWindowsEx. Useful for testing.
 * @return TRUE on success (or dry-run), FALSE on failure.
 */
__declspec(dllexport) BOOL ShutdownPowerOff(BOOL dry_run) {
    if (dry_run)
        return TRUE;
    if (!enable_shutdown_privilege())
        return FALSE;
    return ExitWindowsEx(EXIT_WINDOWS_POWEROFF | EXIT_WINDOWS_FORCEHUNG, SHUTDOWN_REASON_PLANNED);
}

#endif /* POWEROFF_C */
