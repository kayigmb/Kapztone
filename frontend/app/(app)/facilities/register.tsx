"use client";

import { Combobox } from "@base-ui/react/combobox";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { MapPin, Undo2, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";

import { PageHeader } from "@/components/chaintrack/page-header";
import { useRecordTitle } from "@/components/chaintrack/record-title";
import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { MapView } from "@/components/chaintrack/map";
import { DataWidget } from "@/components/chaintrack/data-widget";
import { FieldNote } from "@/components/chaintrack/field-note";
import { Notice } from "@/components/chaintrack/notice";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { endpoints, problemOf } from "@/lib/api";
import { capacityPeriods, copy, queryKeys, routes, upload } from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import type { PlaceSuggestion } from "@/lib/geocode";
import { centreOf, toGeometry, type DrawMode, type LonLat } from "@/lib/geo";
import {
  facilityFormSchema,
  type FacilityForm,
  type FacilityFormInput,
} from "@/lib/schemas/facility";
import { cn } from "@/lib/utils";

const t = copy.facilities.new;
const periodItems = capacityPeriods.map((p) => ({
  value: p.value,
  label: p.label,
}));

export function Register() {
  useRecordTitle(t.title);
  const router = useRouter();
  const client = useQueryClient();
  const kinds = useQuery({
    queryKey: queryKeys.facilityKinds,
    queryFn: endpoints.facilityKinds,
  });
  const [locating, setLocating] = useState(false);
  const [searchText, setSearchText] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [mapFocus, setMapFocus] = useState<LonLat>();
  const [mode, setMode] = useState<DrawMode>("pin");
  const [points, setPoints] = useState<LonLat[]>([]);
  const [licence, setLicence] = useState<File>();
  const {
    register,
    control,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<FacilityFormInput, unknown, FacilityForm>({
    ...liveValidation,
    resolver: zodResolver(facilityFormSchema),
    defaultValues: {
      name: "",
      kind: "",
      licence_number: "",
      licence_valid_to: "",
      capacity_unit: "kg",
      capacity_period: "season",
    },
  });

  useEffect(() => {
    const timer = window.setTimeout(
      () => setSearchTerm(searchText.trim()),
      300,
    );
    return () => window.clearTimeout(timer);
  }, [searchText]);

  const placeSearch = useQuery({
    queryKey: queryKeys.geocode(searchTerm),
    queryFn: ({ signal }) => endpoints.searchPlaces(searchTerm, signal),
    enabled: searchTerm.length >= 3,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  const geometry = toGeometry(mode, points);
  const place = (next: LonLat[], m = mode) => {
    setPoints(next);
    setValue("location", toGeometry(m, next) as FacilityFormInput["location"], {
      shouldValidate: next.length > 0,
    });
  };
  const centre = geometry && centreOf([geometry]);

  const save = useMutation({
    mutationFn: async (v: FacilityForm) => {
      const f = await endpoints.createFacility({
        name: v.name,
        kind: v.kind,
        licence_number: v.licence_number,
        licence_valid_to: v.licence_valid_to,
        capacity: v.capacity,
        capacity_unit: v.capacity_unit,
        capacity_period: v.capacity_period,
        location: v.location,
      });
      // Two calls (the API takes the licence as its own upload); a failed upload leaves the facility without one.
      if (licence)
        await endpoints
          .uploadLicence(f.data.id, licence)
          .catch(() => toast.error(t.licenceFailed));
      return f;
    },
    onSuccess: () => {
      client.invalidateQueries({ queryKey: queryKeys.facilities });
      router.push(routes.facilities);
    },
  });

  const locate = () => {
    if (!navigator.geolocation) {
      toast.error(t.gpsFail);
      return;
    }
    setLocating(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const p: LonLat = [
          Number(pos.coords.longitude.toFixed(6)),
          Number(pos.coords.latitude.toFixed(6)),
        ];
        setMode("pin");
        place([p], "pin");
        setLocating(false);
      },
      () => {
        toast.error(t.gpsFail);
        setLocating(false);
      },
      { enableHighAccuracy: true, timeout: 10_000 },
    );
  };

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title} sub={t.sub} />

      <form
        onSubmit={handleSubmit((v) => save.mutate(v))}
        noValidate
        className="grid gap-6 lg:grid-cols-[minmax(0,1.2fr)_minmax(0,1fr)] lg:items-start"
      >
        <section className="grid min-w-0 gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
          <h2 className="text-title">{t.steps[0]}</h2>
          <div
            role="group"
            aria-label={t.steps[0]}
            className="flex flex-wrap gap-2"
          >
            {(["pin", "area"] as const).map((m) => (
              <button
                key={m}
                type="button"
                aria-pressed={mode === m}
                onClick={() => {
                  setMode(m);
                  place([], m);
                }}
                className={cn(
                  "min-h-tap rounded-full border px-4 text-sm font-semibold",
                  mode === m
                    ? "border-primary bg-primary-soft text-primary"
                    : "border-border-strong bg-surface text-ink-muted hover:bg-surface-sunken",
                )}
              >
                {t.modes[m]}
              </button>
            ))}
          </div>
          <p className="text-small text-ink-muted">{t.hint[mode]}</p>
          <Combobox.Root
            items={placeSearch.data ?? []}
            filter={null}
            inputValue={searchText}
            onInputValueChange={setSearchText}
            itemToStringLabel={(p: PlaceSuggestion) => p.label}
            onValueChange={(p: PlaceSuggestion | null) => {
              if (!p) return;
              setMode("pin");
              place([p.coordinates], "pin");
              setMapFocus(p.coordinates);
              setSearchText("");
            }}
          >
            <div className="grid gap-1.5">
              <label htmlFor="place-search" className="text-sm font-semibold">
                {t.search}
              </label>
              <Combobox.Input
                id="place-search"
                autoComplete="off"
                aria-describedby="place-search-note"
                render={<Input />}
              />
              <FieldNote id="place-search-note" help={t.searchHelp} />
            </div>
            {searchTerm.length >= 3 && (
              <Combobox.Portal>
                <Combobox.Positioner sideOffset={4} className="z-50">
                  <Combobox.Popup
                    aria-label={t.searchResults}
                    className="max-h-72 w-(--anchor-width) overflow-y-auto rounded-md border border-border bg-surface p-1 shadow-card"
                  >
                    <Combobox.Status className="px-3 py-2 text-small text-ink-muted empty:hidden">
                      {placeSearch.isFetching
                        ? copy.loading
                        : placeSearch.isError
                          ? problemOf(placeSearch.error)
                          : null}
                    </Combobox.Status>
                    {!placeSearch.isFetching && !placeSearch.isError && (
                      <Combobox.Empty className="px-3 py-2 text-small text-ink-muted empty:hidden">
                        {t.searchNone}
                      </Combobox.Empty>
                    )}
                    <Combobox.List>
                      {(p: PlaceSuggestion) => (
                        <Combobox.Item
                          key={`${p.label}-${p.coordinates.join()}`}
                          value={p}
                          className="flex min-h-tap cursor-pointer items-center gap-2 rounded-sm px-3 py-2 text-sm data-highlighted:bg-primary-soft"
                        >
                          <MapPin
                            className="size-4 shrink-0 text-primary"
                            aria-hidden
                          />
                          {p.label}
                        </Combobox.Item>
                      )}
                    </Combobox.List>
                  </Combobox.Popup>
                </Combobox.Positioner>
              </Combobox.Portal>
            )}
          </Combobox.Root>
          <MapView
            label={t.mapLabel}
            className="h-80 overflow-hidden rounded-md border border-border lg:h-96"
            draft={{ points, geometry }}
            focus={mapFocus}
            onPick={(p) => place(mode === "pin" ? [p] : [...points, p])}
          />
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p
              className="font-mono text-small text-ink-muted"
              aria-live="polite"
            >
              {mode === "area" && <>{t.corners(points.length)} · </>}
              {centre
                ? `${t.centre} ${centre[1].toFixed(5)}, ${centre[0].toFixed(5)}`
                : null}
            </p>
            <div className="flex flex-wrap gap-2">
              {mode === "area" && (
                <Button
                  type="button"
                  variant="ghost"
                  disabled={!points.length}
                  onClick={() => place(points.slice(0, -1))}
                >
                  <Undo2 aria-hidden />
                  {t.undo}
                </Button>
              )}
              <Button
                type="button"
                variant="ghost"
                disabled={!points.length}
                onClick={() => place([])}
              >
                <X aria-hidden />
                {t.clear}
              </Button>
            </div>
          </div>
          <FieldNote error={errors.location?.message} />
          <div className="flex flex-wrap items-center gap-3">
            <Button
              type="button"
              variant="secondary"
              disabled={locating}
              onClick={locate}
            >
              <MapPin aria-hidden />
              {t.gps}
            </Button>
            <span className="text-small text-ink-subtle">{t.gpsHelp}</span>
          </div>
        </section>

        <div className="grid min-w-0 content-start gap-6">
          <section className="grid min-w-0 gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
            <h2 className="text-title">{t.steps[1]}</h2>
            <TextField
              label={t.name}
              help={t.nameHelp}
              error={errors.name?.message}
              autoComplete="off"
              {...register("name")}
            />
            <DataWidget query={kinds}>
              {(rows) => (
                <ChoiceCards
                  legend={t.kind}
                  options={rows.map((kind) => ({
                    value: kind.key,
                    label: kind.label,
                    sub: kind.sub,
                  }))}
                  error={errors.kind?.message}
                  {...register("kind")}
                />
              )}
            </DataWidget>
            <TextField
              label={t.licence}
              help={t.licenceHelp}
              error={errors.licence_number?.message}
              mono
              autoComplete="off"
              {...register("licence_number")}
            />
            <TextField
              label={t.licenceValid}
              type="date"
              error={errors.licence_valid_to?.message}
              {...register("licence_valid_to")}
            />
            <label className="grid gap-1.5">
              <span className="text-sm font-semibold">{t.licenceDoc}</span>
              <span className="text-small text-ink-muted">
                {t.licenceDocHelp}
              </span>
              <input
                type="file"
                accept={upload.types.join(",")}
                className="min-h-tap text-sm"
                onChange={(e) => setLicence(e.target.files?.[0])}
              />
            </label>
          </section>

          <section className="grid min-w-0 gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
            <h2 className="text-title">{t.steps[2]}</h2>
            <TextField
              label={t.capacity}
              help={t.capacityHelp}
              error={errors.capacity?.message}
              type="number"
              inputMode="numeric"
              min={1}
              step={1}
              mono
              {...register("capacity")}
            />
            <TextField
              label={t.unit}
              help={t.unitHelp}
              error={errors.capacity_unit?.message}
              mono
              {...register("capacity_unit")}
            />
            <div className="grid gap-1.5">
              <span className="text-sm font-semibold">{t.period}</span>
              <Controller
                control={control}
                name="capacity_period"
                render={({ field }) => (
                  <Select
                    items={periodItems}
                    value={field.value}
                    onValueChange={(v) => field.onChange(v)}
                  >
                    <SelectTrigger
                      aria-label={t.period}
                      className="min-h-tap w-full rounded-md border-border-strong bg-surface text-base"
                    >
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {periodItems.map((p) => (
                        <SelectItem key={p.value} value={p.value}>
                          {p.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldNote error={errors.capacity_period?.message} />
            </div>
          </section>
        </div>

        <div className="grid gap-4 lg:col-span-2">
          <Notice kind="info">{t.approvalNote}</Notice>
          <Button
            type="submit"
            loading={save.isPending}
            loadingText={t.saving}
            className="w-full lg:w-auto lg:justify-self-end"
          >
            {t.save}
          </Button>
        </div>
      </form>
    </main>
  );
}
