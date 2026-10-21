"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { useConfirm } from "@/components/chaintrack/confirm";
import { FieldNote } from "@/components/chaintrack/field-note";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { ApiError, endpoints, problemOf } from "@/lib/api";
import {
  confirmCopy,
  copy,
  orgStatus,
  queryKeys,
  type ReviewDecision,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import { reviewSchema, type Review } from "@/lib/schemas/org";

const t = copy.admin.review;

/** Approve, reject or ask for changes, with a note the organisation will see. */
export function ReviewForm({ orgId }: { orgId: string }) {
  const client = useQueryClient();
  const {
    register,
    handleSubmit,
    setValue,
    setError,
    formState: { errors },
  } = useForm<Review>({
    ...liveValidation,
    resolver: zodResolver(reviewSchema),
    defaultValues: { decision: orgStatus.verified, reason: "" },
  });

  const review = useMutation({
    mutationFn: (r: Review) => endpoints.review(orgId, r),
    meta: { inPlace: true },
    onSuccess: () => client.invalidateQueries({ queryKey: queryKeys.orgs }), // every list, and this organisation
    onError: (err) => {
      if (err instanceof ApiError && err.status === 400)
        return setError(
          "reason",
          { message: err.message },
          { shouldFocus: true },
        );
      toast.error(problemOf(err));
    },
  });

  // Each button submits the same form with its own decision.
  const confirm = useConfirm();
  const decide = (decision: ReviewDecision) => () => {
    setValue("decision", decision);
    handleSubmit(async (r) => {
      const c = confirmCopy.review[decision];
      if (
        await confirm({
          title: c.title,
          body: c.body,
          confirmLabel: c.label,
          danger: decision === orgStatus.rejected,
        })
      )
        review.mutate(r);
    })();
  };
  const noteId = "review-note";

  return (
    <form
      onSubmit={(e) => e.preventDefault()}
      noValidate
      className="grid gap-3 border-t border-border pt-6"
    >
      <h3 className="text-title">{t.title}</h3>
      <div className="grid gap-1.5">
        <label htmlFor="reason" className="text-sm font-semibold">
          {t.note}
        </label>
        <Textarea
          id="reason"
          rows={3}
          aria-invalid={errors.reason ? true : undefined}
          aria-describedby={noteId}
          {...register("reason")}
        />
        <FieldNote
          id={noteId}
          error={errors.reason?.message}
          help={t.noteHelp}
        />
      </div>
      <div className="flex flex-wrap justify-end gap-2">
        <Button
          type="button"
          variant="danger-outline"
          loading={review.isPending}
          onClick={decide(orgStatus.rejected)}
        >
          {t.reject}
        </Button>
        <Button
          type="button"
          variant="secondary"
          loading={review.isPending}
          onClick={decide(orgStatus.changesRequested)}
        >
          {t.changes}
        </Button>
        <Button
          type="button"
          loading={review.isPending}
          loadingText={t.saving}
          onClick={decide(orgStatus.verified)}
        >
          {t.approve}
        </Button>
      </div>
    </form>
  );
}
