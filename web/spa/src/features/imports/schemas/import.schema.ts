import { z } from 'zod'
export const importSchema = z.object({ file: z.instanceof(File).refine((file) => /\.csv$/i.test(file.name), 'validation.csv').refine((file) => file.size > 0 && file.size <= 5 * 1024 * 1024, 'validation.fileSize') })
export type ImportInput = z.infer<typeof importSchema>
