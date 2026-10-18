"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { MapPin, OctagonX, Undo2, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";

import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { MapView } from "@/components/chaintrack/map";
import { Notice } from "@/components/chaintrack/notice";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { endpoints, problemOf } from "@/lib/api";
import { capacityPeriods, copy, facilityKinds, queryKeys, routes, upload } from "@/lib/consts";
import { centreOf, toGeometry, type DrawMode, type LonLat } from "@/lib/geo";
import { facilityFormSchema, type FacilityForm, type FacilityFormInput } from "@/lib/schemas/facility";
import { cn } from "@/lib/utils";

const t = copy.facilities.new;
const periodItems = capacityPeriods.map((p) => ({ value: p.value, label: p.label }));

export function Register() {
  const router = useRouter();
  const client = useQueryClient();
  const [locating, setLocating] = useState(false);
  const [mode, setMode] = useState<DrawMode>("pin");
  const [points, setPoints] = useState<LonLat[]>([]);
  const [licence, setLicence] = useState<File>();
  const { register, control, handleSubmit, setValue, formState: { errors } } = useForm<FacilityFormInput, unknown, FacilityForm>({
    resolver: zodResolver(facilityFormSchema),
    defaultValues: { name: "", kind: "farm_plot", licence_number: "", licence_valid_to: "", capacity_unit: "kg", capacity_period: "season" },
  });

  const geometry = toGeometry(mode, points);
  const place = (next: LonLat[], m = mode) => {
    setPoints(next);
    setValue("location", toGeometry(m, next) as FacilityFormInput["location"], { shouldValidate: next.length > 0 });
  };
  const centre = geometry && centreOf([geometry]);

  const save = useMutation({
    mutationFn: async (v: FacilityForm) => {
      const f = await endpoints.createFacility({
        name: v.name, kind: v.kind, licence_number: v.licence_number, licence_valid_to: v.licence_valid_to,
        capacity: v.capacity, capacity_unit: v.capacity_unit, capacity_period: v.capacity_period, location: v.location,
      });
      // ponytail: two calls (the API takes the licence as its own upload); a failed upload leaves the facility without one.
      if (licence) await endpoints.uploadLicence(f.id, licence).catch(() => toast.error(t.licenceFailed));
      return f;
    },
    onSuccess: (_r, v) => {
      toast.success(t.created(v.name));
      client.invalidateQueries({ queryKey: queryKeys.facilities });
      router.push(routes.facilities);
    },
    onError: (err) => toast.error(problemOf(err)),
  });

  const locate = () => {
    if (!navigator.geolocation) { toast.error(t.gpsFail); return; }
    setLocating(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const p: LonLat = [Number(pos.coords.longitude.toFixed(6)), Number(pos.coords.latitude.toFixed(6))];
        setMode("pin");
        place([p], "pin");
        setLocating(false);
      },
      () => { toast.error(t.gpsFail); setLocating(false); },
      { enableHighAccuracy: true, timeout: 10_000 },
    );
  };

  return (
    <main className="grid content-start gap-6 p-4 lg:p-8">
      <header className="grid gap-1">
        <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{t.sub}</p>
      </header>

      <form onSubmit={handleSubmit((v) => save.mutate(v))} noValidate className="grid max-w-2xl gap-6">
        <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
          <h2 className="text-title">{t.steps[0]}</h2>
          <div role="group" aria-label={t.steps[0]} className="flex flex-wrap gap-2">
            {(["pin", "area"] as const).map((m) => (
              <button
                key={m} type="button" aria-pressed={mode === m} onClick={() => { setMode(m); place([], m); }}
                className={cn("min-h-tap rounded-full border px-4 text-sm font-semibold",
                  mode === m ? "border-primary bg-primary-soft text-primary" : "border-border-strong bg-surface text-ink-muted hover:bg-surface-sunken")}
              >
                {t.modes[m]}
              </button>
            ))}
          </div>
          <p className="text-small text-ink-muted">{t.hint[mode]}</p>
          <MapView
            label={t.mapLabel} className="h-80 overflow-hidden rounded-md border border-border lg:h-96"
            draft={{ points, geometry }} onPick={(p) => place(mode === "pin" ? [p] : [...points, p])}
          />
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="font-mono text-small text-ink-muted" aria-live="polite">
              {mode === "area" && <>{t.corners(points.length)} · </>}
              {centre ? `${t.centre} ${centre[1].toFixed(5)}, ${centre[0].toFixed(5)}` : null}
            </p>
            <div className="flex flex-wrap gap-2">
              {mode === "area" && (
                <Button type="button" variant="ghost" disabled={!points.length} onClick={() => place(points.slice(0, -1))}>
                  <Undo2 aria-hidden />
                  {t.undo}
                </Button>
              )}
              <Button type="button" variant="ghost" disabled={!points.length} onClick={() => place([])}>
                <X aria-hidden />
                {t.clear}
              </Button>
            </div>
          </div>
          {errors.location && (
            <p role="alert" className="flex gap-1.5 text-small font-semibold text-danger">
              <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
              {errors.location.message}
            </p>
          )}
          <div className="flex flex-wrap items-center gap-3">
            <Button type="button" variant="secondary" disabled={locating} onClick={locate}>
              <MapPin aria-hidden />
              {t.gps}
            </Button>
            <span className="text-small text-ink-subtle">{t.gpsHelp}</span>
          </div>
        </section>

        <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
          <h2 className="text-title">{t.steps[1]}</h2>
          <TextField label={t.name} help={t.nameHelp} error={errors.name?.message} autoComplete="off" {...register("name")} />
          <ChoiceCards legend={t.kind} options={facilityKinds} error={errors.kind?.message} {...register("kind")} />
          <TextField label={t.licence} help={t.licenceHelp} error={errors.licence_number?.message} mono autoComplete="off" {...register("licence_number")} />
          <TextField label={t.licenceValid} type="date" error={errors.licence_valid_to?.message} {...register("licence_valid_to")} />
          <label className="grid gap-1.5">
            <span className="text-sm font-semibold">{t.licenceDoc}</span>
            <span className="text-small text-ink-muted">{t.licenceDocHelp}</span>
            <input type="file" accept={upload.types.join(",")} className="min-h-tap text-sm"
              onChange={(e) => setLicence(e.target.files?.[0])} />
          </label>
        </section>

        <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
          <h2 className="text-title">{t.steps[2]}</h2>
          <TextField
            label={t.capacity} help={t.capacityHelp} error={errors.capacity?.message}
            type="number" inputMode="numeric" min={1} step={1} mono {...register("capacity")}
          />
          <TextField label={t.unit} help={t.unitHelp} error={errors.capacity_unit?.message} mono {...register("capacity_unit")} />
          <div className="grid gap-1.5">
            <span className="text-sm font-semibold">{t.period}</span>
            <Controller control={control} name="capacity_period" render={({ field }) => (
              <Select items={periodItems} value={field.value} onValueChange={(v) => field.onChange(v)}>
                <SelectTrigger aria-label={t.period} className="min-h-tap w-full rounded-md border-border-strong bg-surface text-base">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>{periodItems.map((p) => <SelectItem key={p.value} value={p.value}>{p.label}</SelectItem>)}</SelectContent>
              </Select>
            )} />
            {errors.capacity_period && (
              <p className="flex gap-1.5 text-small font-semibold text-danger">
                <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
                {errors.capacity_period.message}
              </p>
            )}
          </div>
        </section>

        <Notice kind="info">{t.approvalNote}</Notice>
        <Button type="submit" size="scan" disabled={save.isPending} className="w-full lg:w-auto">
          {save.isPending ? t.saving : t.save}
        </Button>
      </form>
    </main>
  );
}
