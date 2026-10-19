// How every form validates (react-hook-form + zod): a field shows its error
// once the person leaves it, then updates on every keystroke until it's right.

export const liveValidation = {
  mode: "onTouched",
  reValidateMode: "onChange",
} as const;
