/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useEffect, useMemo, useState } from 'react'
import { Search, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { searchUsers } from '@/features/users/api'
import type { User } from '@/features/users/types'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

const OPTION_KEY = 'UsageLeaderboardIgnoredUserIds'

type LeaderboardSettingsSectionProps = {
  defaultValue: string
}

function parseIgnoredUserIds(value: string): number[] {
  if (!value.trim()) return []
  try {
    const parsed = JSON.parse(value) as unknown
    if (!Array.isArray(parsed)) return []
    const seen = new Set<number>()
    const ids: number[] = []
    for (const item of parsed) {
      const id = Number(item)
      if (!Number.isInteger(id) || id <= 0 || seen.has(id)) continue
      seen.add(id)
      ids.push(id)
    }
    return ids
  } catch {
    return []
  }
}

function serializeIgnoredUserIds(ids: number[]): string {
  return JSON.stringify(
    [...new Set(ids.filter((id) => Number.isInteger(id) && id > 0))]
  )
}

function userLabel(user: User): string {
  const name = user.display_name || user.username || `User #${user.id}`
  const email = user.email ? ` · ${user.email}` : ''
  return `${name} (#${user.id})${email}`
}

export function LeaderboardSettingsSection({
  defaultValue,
}: LeaderboardSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const initialIds = useMemo(
    () => parseIgnoredUserIds(defaultValue),
    [defaultValue]
  )
  const [selectedIds, setSelectedIds] = useState<number[]>(initialIds)
  const [searchText, setSearchText] = useState('')
  const [searching, setSearching] = useState(false)
  const [results, setResults] = useState<User[]>([])
  const [knownUsers, setKnownUsers] = useState<Record<number, User>>({})

  useEffect(() => {
    setSelectedIds(initialIds)
  }, [initialIds])

  const serialized = serializeIgnoredUserIds(selectedIds)
  const initialSerialized = serializeIgnoredUserIds(initialIds)
  const hasChanges = serialized !== initialSerialized

  const handleSearch = async () => {
    const keyword = searchText.trim()
    if (!keyword) {
      setResults([])
      return
    }
    setSearching(true)
    try {
      const response = await searchUsers({
        keyword,
        p: 1,
        page_size: 8,
      })
      if (!response.success) {
        toast.error(response.message || t('Failed to search users'))
        return
      }
      const users = response.data?.items ?? []
      setResults(users)
      setKnownUsers((prev) => {
        const next = { ...prev }
        for (const user of users) {
          next[user.id] = user
        }
        return next
      })
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : t('Failed to search users')
      )
    } finally {
      setSearching(false)
    }
  }

  const addUser = (user: User) => {
    setKnownUsers((prev) => ({ ...prev, [user.id]: user }))
    setSelectedIds((prev) =>
      prev.includes(user.id) ? prev : [...prev, user.id]
    )
  }

  const removeUser = (id: number) => {
    setSelectedIds((prev) => prev.filter((item) => item !== id))
  }

  const reset = () => {
    setSelectedIds(initialIds)
  }

  const save = async () => {
    await updateOption.mutateAsync({
      key: OPTION_KEY,
      value: serialized,
    })
  }

  return (
    <SettingsSection title={t('Leaderboard settings')}>
      <SettingsPageFormActions
        onSave={save}
        onReset={reset}
        isSaving={updateOption.isPending}
        isSaveDisabled={!hasChanges}
        isResetDisabled={!hasChanges}
        saveLabel='Save leaderboard settings'
      />

      <div className='space-y-4'>
        <div className='flex gap-2'>
          <Input
            value={searchText}
            onChange={(event) => setSearchText(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                event.preventDefault()
                void handleSearch()
              }
            }}
            placeholder={t('Search users by name, email, or ID')}
          />
          <Button
            type='button'
            variant='outline'
            onClick={() => void handleSearch()}
            disabled={searching}
          >
            <Search data-icon='inline-start' />
            {searching ? t('Searching...') : t('Search')}
          </Button>
        </div>

        {results.length > 0 && (
          <div className='border-border divide-border rounded-lg border'>
            {results.map((user) => {
              const selected = selectedIds.includes(user.id)
              return (
                <button
                  key={user.id}
                  type='button'
                  onClick={() => addUser(user)}
                  disabled={selected}
                  className='hover:bg-muted/60 flex w-full items-center justify-between gap-3 px-3 py-2 text-left text-sm disabled:cursor-default disabled:opacity-60'
                >
                  <span className='min-w-0 truncate'>{userLabel(user)}</span>
                  <span className='text-muted-foreground shrink-0 text-xs'>
                    {selected ? t('Selected') : t('Add')}
                  </span>
                </button>
              )
            })}
          </div>
        )}

        <div className='space-y-2'>
          <div className='text-sm font-medium'>{t('Ignored users')}</div>
          {selectedIds.length === 0 ? (
            <div className='border-border text-muted-foreground rounded-lg border border-dashed px-3 py-6 text-center text-sm'>
              {t('No ignored users')}
            </div>
          ) : (
            <div className='flex flex-wrap gap-2'>
              {selectedIds.map((id) => {
                const user = knownUsers[id]
                return (
                  <Badge key={id} variant='outline' className='h-7 gap-1.5'>
                    <span>{user ? userLabel(user) : `User #${id}`}</span>
                    <button
                      type='button'
                      onClick={() => removeUser(id)}
                      className='hover:text-destructive rounded-sm'
                      aria-label={t('Remove user')}
                    >
                      <X className='size-3' />
                    </button>
                  </Badge>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </SettingsSection>
  )
}
