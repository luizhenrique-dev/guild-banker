import { z } from 'zod'

export const createGuildSchema = z.object({
  display_name: z.string().trim().min(1, 'onboarding.nameRequired').refine((name) => [...name].length <= 70, 'onboarding.nameLength'),
  description: z.string(),
})
export type CreateGuildInput = z.infer<typeof createGuildSchema>
