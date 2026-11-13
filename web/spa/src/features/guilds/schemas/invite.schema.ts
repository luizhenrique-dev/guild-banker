import { z } from 'zod'
export const maxInviteEmailLength = 254
export const inviteEmailPattern = /^[a-z0-9_%+\-]+(?:\.[a-z0-9_%+\-]+)*@(?:[a-z0-9](?:[a-z0-9\-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$/
export const inviteSchema = z.object({
  email: z.string().trim().toLowerCase().max(maxInviteEmailLength, 'validation.email').regex(inviteEmailPattern, 'validation.email'),
})
export type InviteInput = z.infer<typeof inviteSchema>
