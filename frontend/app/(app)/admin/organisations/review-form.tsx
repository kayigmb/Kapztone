"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { OctagonX } from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { ApiError, endpoints, problemOf } from "@/lib/api";
import { copy, orgStatus, queryKeys, type ReviewDecision } from "@/lib/consts";
import { reviewSchema, type Review } from "@/lib/schemas/org";

const t = copy.admin.review;

/** Approve, reject or ask for changes, with a note the organisation will see. */
export function ReviewForm({ orgId }: { orgId: string }) {
  const client = useQueryClient();
  const { register, handleSubmit, setValue, setError, formState: { errors } } = useForm<Review>({
    resolver: zodResolver(reviewSchema),
    defaultValues: { decision: orgStatus.verified, reason: "" },
  });

  const review = useMutation({
    mutationFn: (r: Review) => endpoints.review(orgId, r),
    onSuccess: (_, r) => {
      toast.success(t.done[r.decision]);
      return client.invalidateQueries({ queryKey: queryKeys.orgs }); // every list, and this organisation
    },
    onError: (err) => {
      if (err instanceof ApiError && err.status === 409) return toast.error(t.noDocuments);
      if (err instanceof ApiError && err.status === 400) return setError("reason", { message: err.message }, { shouldFocus: true });
      toast.error(problemOf(err));
    },
  });

  // Each button submits the same form with its own decision.
  const decide = (decision: ReviewDecision) => () => {
    setValue("decision", decision);
    handleSubmit((r) => review.mutate(r))();
  };
  const noteId = "review-note";

  return (
    <form onSubmit={(e) => e.preventDefault()} noValidate className="grid gap-3 border-t border-border pt-6">
      <h3 className="text-title">{t.title}</h3>
      <div className="grid gap-1.5">
        <label htmlFor="reason" className="text-sm font-semibold">{t.note}</label>
        <Textarea id="reason" rows={3} aria-invalid={errors.reason ? true : undefined} aria-describedby={noteId}
          {...register("reason")} />
        {errors.reason ? (
          <p id={noteId} className="flex gap-1.5 text-small font-semibold text-danger">
            <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
            {errors.reason.message}
          </p>
        ) : (
          <p id={noteId} className="text-small text-ink-subtle">{t.noteHelp}</p>
        )}
      </div>
      <div className="flex flex-wrap justify-end gap-2">
        <Button type="button" variant="danger-outline" disabled={review.isPending} onClick={decide(orgStatus.rejected)}>{t.reject}</Button>
        <Button type="button" variant="secondary" disabled={review.isPending} onClick={decide(orgStatus.changesRequested)}>{t.changes}</Button>
        <Button type="button" disabled={review.isPending} onClick={decide(orgStatus.verified)}>
          {review.isPending ? t.saving : t.approve}
        </Button>
      </div>
    </form>
  );
}
