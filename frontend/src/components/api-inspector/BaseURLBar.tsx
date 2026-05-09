import { useEffect, useMemo, useRef, useState } from "react";
import {
  Globe,
  Check,
  X,
  LogIn,
  LogOut,
  ChevronsUpDown,
  Settings2,
  FolderKanban,
} from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { useProjectStore } from "@/store/projectStore";
import { useEndpointsStore } from "@/store/endpointsStore";
import type { ScannedEndpoint } from "@/services/scannerService";
import { useHttpMethod } from "@/hooks/useHttpMethod";
import { cn } from "@/lib/utils";
import { EnvironmentSwitcher } from "./EnvironmentSwitcher";
import { CapturedValuesPopover } from "./CapturedValuesPopover";
import { AccountSelector } from "@/components/accounts/AccountSelector";
import { MockToggle } from "@/components/mock/MockToggle";
import { useUIStore } from "@/store/uiStore";
import { useCollectionsStore } from "@/store/collectionsStore";
import { VarInput } from "./VarInput";
import { VarText } from "./VarText";
import { useEnvironmentStore } from "@/store/environmentStore";
import { useCapturesStore } from "@/store/capturesStore";

const EMPTY_ENDPOINTS: ScannedEndpoint[] = [];
const EMPTY_CAPTURED: import('@/services/capturesService').CapturedValue[] = [];

export function BaseURLBar() {
  const activeProjectId = useProjectStore((s) => s.activeProjectId);
  const projects = useProjectStore((s) => s.projects);
  const updateBaseURL = useProjectStore((s) => s.updateBaseURL);
  const updateAuthRoutes = useProjectStore((s) => s.updateAuthRoutes);
  const activeTabId = useUIStore((s) =>
    activeProjectId ? s.activeInspectorTabByProject[activeProjectId] ?? null : null,
  );
  const project = projects.find((p) => p.id === activeProjectId);
  const allEndpoints = useEndpointsStore((s) =>
    activeProjectId
      ? (s.byProject[activeProjectId] ?? EMPTY_ENDPOINTS)
      : EMPTY_ENDPOINTS,
  );
  const capturedValues = useCapturesStore((s) => activeProjectId ? s.byProject[activeProjectId] ?? EMPTY_CAPTURED : EMPTY_CAPTURED)
  const envs = useEnvironmentStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? null : null,
  );
  const activeEnv = envs?.find((e) => e.id === project?.activeEnvironmentId) ?? null;
  const variableNames = activeEnv?.vars ?? {};

  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState(project?.baseUrl ?? "");
  const [busy, setBusy] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setValue(project?.baseUrl ?? "");
  }, [project?.id, project?.baseUrl]);

  if (!project) return null;

  const beginEdit = () => {
    setEditing(true);
    setTimeout(() => inputRef.current?.focus(), 0);
  };

  const cancel = () => {
    setValue(project.baseUrl ?? "");
    setEditing(false);
  };

  const save = async () => {
    const trimmed = value.trim();
    if (!trimmed || trimmed === project.baseUrl) {
      setEditing(false);
      return;
    }
    setBusy(true);
    try {
      await updateBaseURL(project.id, trimmed);
      setEditing(false);
    } finally {
      setBusy(false);
    }
  };

  const onKey = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      void save();
    } else if (e.key === "Escape") {
      e.preventDefault();
      cancel();
    }
  };

  return (
    <div className="h-10.5 px-4 border-b border-border/50 flex items-center gap-2 bg-transparent">
      <Globe className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
      <span className="text-[10px] uppercase tracking-wider text-muted-foreground shrink-0">
        Base URL
      </span>
      {editing ? (
        <div className="flex items-center gap-1.5 flex-1 min-w-0">
          <VarInput
            ref={inputRef}
            value={value}
            onChange={(v) => setValue(v)}
            onKeyDown={onKey}
            placeholder="http://localhost:8000"
            className="h-7 text-[12px] font-mono"
            disabled={busy}
            variables={variableNames}
          />
          <Button
            size="sm"
            variant="outline"
            className="h-7 px-2"
            onClick={save}
            disabled={busy}
          >
            <Check className="w-3.5 h-3.5" />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            className="h-7 px-2"
            onClick={cancel}
            disabled={busy}
          >
            <X className="w-3.5 h-3.5" />
          </Button>
        </div>
      ) : (
        <button
          type="button"
          onClick={beginEdit}
          className="flex-1 min-w-0 text-left font-mono text-[12px] text-foreground/85 hover:text-foreground truncate"
          title={project.baseUrl || "Click to set base URL"}
        >
          {project.baseUrl ? (
            <VarText value={project.baseUrl} variables={variableNames} />
          ) : (
            <span className="text-muted-foreground italic">
              click to set base URL
            </span>
          )}
        </button>
      )}

      <MockToggle projectId={project.id} />
      <AccountSelector projectId={project.id} tabId={activeTabId} />
      <EnvironmentSwitcher />
      <CapturedValuesPopover
        projectId={activeProjectId ?? null}
        values={capturedValues}
        onChange={(vals) => activeProjectId && useCapturesStore.getState().set(activeProjectId, vals)}
      />
      <CollectionsButton />
      <AuthRoutesPopover
        projectId={project.id}
        endpoints={allEndpoints}
        loginEndpointId={project.loginEndpointId ?? ""}
        loginTokenPath={project.loginTokenPath ?? ""}
        logoutEndpointId={project.logoutEndpointId ?? ""}
        onSave={updateAuthRoutes}
      />
    </div>
  );
}

interface AuthRoutesPopoverProps {
  projectId: string;
  endpoints: ScannedEndpoint[];
  loginEndpointId: string;
  loginTokenPath: string;
  logoutEndpointId: string;
  onSave: (
    id: string,
    loginId: string,
    logoutId: string,
    tokenPath: string,
  ) => Promise<void>;
}

function AuthRoutesPopover({
  projectId,
  endpoints,
  loginEndpointId,
  loginTokenPath,
  logoutEndpointId,
  onSave,
}: AuthRoutesPopoverProps) {
  const [open, setOpen] = useState(false);
  const [loginDraft, setLoginDraft] = useState(loginEndpointId);
  const [logoutDraft, setLogoutDraft] = useState(logoutEndpointId);
  const [pathDraft, setPathDraft] = useState(loginTokenPath);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setLoginDraft(loginEndpointId);
    setLogoutDraft(logoutEndpointId);
    setPathDraft(loginTokenPath);
  }, [loginEndpointId, logoutEndpointId, loginTokenPath, projectId]);

  const writeEndpoints = useMemo(
    () =>
      endpoints.filter((e) => {
        const m = e.method.toUpperCase();
        return m === "POST" || m === "PUT" || m === "PATCH" || m === "DELETE";
      }),
    [endpoints],
  );

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(projectId, loginDraft, logoutDraft, pathDraft.trim());
      setOpen(false);
    } finally {
      setSaving(false);
    }
  };

  const configured = Boolean(loginEndpointId || logoutEndpointId);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            "shrink-0 h-7 inline-flex items-center gap-1.5 px-2 rounded-md text-[11px] border border-border/50 hover:bg-accent/40 transition-colors",
            configured ? "text-foreground" : "text-muted-foreground",
          )}
          title="Configure auth routes"
        >
          <Settings2 className="w-3 h-3" />
          <span>Auth routes</span>
          {configured && (
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[380px] p-3 space-y-3">
        <RouteSelect
          icon={LogIn}
          label="Login"
          endpoints={writeEndpoints}
          value={loginDraft}
          onChange={setLoginDraft}
          placeholder="Select login route"
        />
        <RouteSelect
          icon={LogOut}
          label="Logout"
          endpoints={writeEndpoints}
          value={logoutDraft}
          onChange={setLogoutDraft}
          placeholder="Select logout route"
        />
        <div className="space-y-1.5">
          <label className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            Token JSONPath
          </label>
          <Input
            value={pathDraft}
            onChange={(e) => setPathDraft(e.target.value)}
            placeholder="data.token (auto-detected if empty)"
            className="h-7 text-[12px] font-mono"
          />
        </div>
        <Button
          size="sm"
          onClick={handleSave}
          disabled={saving}
          className="w-full h-7 text-[11px]"
        >
          {saving ? "Saving..." : "Save"}
        </Button>
      </PopoverContent>
    </Popover>
  );
}

interface RouteSelectProps {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  endpoints: ScannedEndpoint[];
  value: string;
  onChange: (id: string) => void;
  placeholder: string;
}

function RouteSelect({
  icon: Icon,
  label,
  endpoints,
  value,
  onChange,
  placeholder,
}: RouteSelectProps) {
  const { getMethodColor } = useHttpMethod();
  const [open, setOpen] = useState(false);
  const selected = endpoints.find((e) => e.id === value);

  return (
    <div className="space-y-1.5">
      <label className="flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        <Icon className="w-3 h-3" />
        {label}
      </label>
      <Popover open={open} onOpenChange={setOpen} modal>
        <PopoverTrigger asChild>
          <button
            type="button"
            className="w-full h-8 px-2 inline-flex items-center gap-2 rounded-md border border-border/40 bg-muted/40 text-[12px] hover:bg-accent/40 transition-colors"
          >
            {selected ? (
              <>
                <span
                  className={cn(
                    "inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5",
                    getMethodColor(selected.method),
                  )}
                >
                  {selected.method}
                </span>
                <span className="font-mono truncate flex-1 text-left">
                  {selected.path}
                </span>
              </>
            ) : (
              <span className="flex-1 text-left text-muted-foreground italic">
                {placeholder}
              </span>
            )}
            <ChevronsUpDown className="w-3 h-3 opacity-60 shrink-0" />
          </button>
        </PopoverTrigger>
        <PopoverContent align="start" className="w-[360px] p-0">
          <Command
            filter={(itemValue, search) => {
              const v = itemValue.toLowerCase();
              return v.includes(search.toLowerCase()) ? 1 : 0;
            }}
          >
            <CommandInput
              placeholder="Search endpoints..."
              className="h-8 text-[12px]"
            />
            <CommandList className="max-h-64">
              <CommandEmpty className="py-4 text-center text-[11.5px] text-muted-foreground">
                No endpoints
              </CommandEmpty>
              <CommandGroup>
                <CommandItem
                  value="__none__"
                  onSelect={() => {
                    onChange("");
                    setOpen(false);
                  }}
                  className="text-[11.5px] italic text-muted-foreground"
                >
                  — None —
                </CommandItem>
                {endpoints.map((ep) => (
                  <CommandItem
                    key={ep.id}
                    value={`${ep.method} ${ep.path}`}
                    onSelect={() => {
                      onChange(ep.id);
                      setOpen(false);
                    }}
                    className="gap-2 text-[11.5px]"
                  >
                    <span
                      className={cn(
                        "inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5",
                        getMethodColor(ep.method),
                      )}
                    >
                      {ep.method}
                    </span>
                    <span className="font-mono truncate flex-1">{ep.path}</span>
                    {value === ep.id && (
                      <Check className="w-3 h-3 text-primary" />
                    )}
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  );
}

function CollectionsButton() {
  const projectId = useProjectStore((s) => s.activeProjectId);
  const setCurrentPage = useUIStore((s) => s.setCurrentPage);
  const count = useCollectionsStore((s) =>
    projectId ? (s.byProject[projectId]?.length ?? 0) : 0,
  );
  return (
    <button
      type="button"
      onClick={() => setCurrentPage("collections")}
      className="inline-flex items-center gap-1.5 h-7 px-2 rounded-md border border-border/50 bg-card text-[11px] text-muted-foreground hover:text-foreground hover:bg-accent/60 transition-colors"
      title="Open Collections page"
    >
      <FolderKanban className="w-3 h-3" />
      <span>Collections</span>
      <span className="font-mono text-[10.5px] tabular-nums text-muted-foreground">
        {count}
      </span>
    </button>
  );
}
