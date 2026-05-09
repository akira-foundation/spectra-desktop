import {
  ListCollections,
  SaveCollection,
  DeleteCollection,
  RunCollection,
} from '../../wailsjs/go/app/App'
import { app } from '../../wailsjs/go/models'

export type Collection = app.CollectionDTO
export type CollectionItem = app.CollectionItemDTO
export type CollectionRun = app.CollectionRunDTO
export type CollectionRunItem = app.CollectionRunItemDTO

export interface SaveCollectionInput {
  id?: string
  projectID: string
  name: string
  description?: string
  sortOrder?: number
  items: CollectionItem[]
}

export const collectionsService = {
  async list(projectId: string): Promise<Collection[]> {
    const rows = await ListCollections(projectId)
    return rows ?? []
  },
  async save(input: SaveCollectionInput): Promise<Collection | null> {
    const result = await SaveCollection(
      app.SaveCollectionInput.createFrom({
        id: input.id ?? '',
        projectID: input.projectID,
        name: input.name,
        description: input.description ?? '',
        sortOrder: input.sortOrder ?? 0,
        items: input.items,
      }),
    )
    return result ?? null
  },
  async remove(id: string): Promise<void> {
    await DeleteCollection(id)
  },
  async run(id: string): Promise<CollectionRun | null> {
    return (await RunCollection(id)) ?? null
  },
}
