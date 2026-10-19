"use client";

import { Camera, MapPin, X } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { copy as ui, maxHandoverPhotos, upload } from "@/lib/consts";

export type Place = { lat: number; lon: number } | undefined;

/**
 * Where it happened (the device's GPS, optional) and photos (optional): both
 * end up in the signed record. Used when sending and when receiving a lot.
 */
export function PlaceAndPhotos({
  place,
  onPlace,
  photos,
  onPhotos,
  copy,
}: {
  place: Place;
  onPlace: (p: Place) => void;
  photos: File[];
  onPhotos: (f: File[]) => void;
  copy: {
    location: string;
    locationAdded: (lat: string, lon: string) => string;
    locationFailed: string;
    photos: string;
    photosHelp: (max: number) => string;
  };
}) {
  const [locating, setLocating] = useState(false);
  const [failed, setFailed] = useState(false);
  const locate = () => {
    if (!navigator.geolocation) return setFailed(true);
    setLocating(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLocating(false);
        setFailed(false);
        onPlace({ lat: pos.coords.latitude, lon: pos.coords.longitude });
      },
      () => {
        setLocating(false);
        setFailed(true);
      },
      { enableHighAccuracy: true, timeout: 15_000 },
    );
  };

  return (
    <div className="grid gap-4">
      <div className="grid gap-1.5">
        {place ? (
          <p className="flex min-h-tap items-center gap-2 text-sm font-semibold text-verified-ink">
            <MapPin className="size-4 shrink-0" aria-hidden />
            {copy.locationAdded(place.lat.toFixed(5), place.lon.toFixed(5))}
            <button
              type="button"
              className="ml-auto grid size-tap place-items-center text-ink-muted"
              onClick={() => onPlace(undefined)}
              aria-label={ui.removeLocation}
            >
              <X className="size-4" aria-hidden />
            </button>
          </p>
        ) : (
          <Button
            type="button"
            variant="secondary"
            onClick={locate}
            disabled={locating}
            className="w-fit"
          >
            <MapPin aria-hidden />
            {copy.location}
          </Button>
        )}
        {failed && (
          <p className="text-small text-ink-muted">{copy.locationFailed}</p>
        )}
      </div>

      <div className="grid gap-2">
        <label className="grid gap-1.5">
          <span className="text-sm font-semibold">{copy.photos}</span>
          <span className="text-small text-ink-muted">
            {copy.photosHelp(maxHandoverPhotos)}
          </span>
          <span className="flex min-h-tap w-fit cursor-pointer items-center gap-2 rounded-md border border-border px-4 text-sm font-semibold hover:border-primary">
            <Camera className="size-5" aria-hidden />
            {copy.photos}
            <input
              type="file"
              accept={upload.imageTypes.join(",")}
              capture="environment"
              multiple
              className="sr-only"
              onChange={(e) => {
                onPhotos(
                  [...photos, ...Array.from(e.target.files ?? [])].slice(
                    0,
                    maxHandoverPhotos,
                  ),
                );
                e.target.value = "";
              }}
            />
          </span>
        </label>
        {photos.length > 0 && (
          <ul className="flex flex-wrap gap-2">
            {photos.map((f, i) => (
              <li
                key={`${f.name}-${i}`}
                className="flex min-h-tap items-center gap-1 rounded-md bg-surface-sunken pl-3 text-small"
              >
                <span className="max-w-[18ch] truncate">{f.name}</span>
                <button
                  type="button"
                  className="grid size-tap place-items-center text-ink-muted"
                  aria-label={ui.removeFile(f.name)}
                  onClick={() => onPhotos(photos.filter((_, j) => j !== i))}
                >
                  <X className="size-4" aria-hidden />
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
