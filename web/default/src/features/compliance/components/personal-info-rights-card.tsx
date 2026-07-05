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
import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { FileText, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  NativeSelect,
  NativeSelectOption,
} from '@/components/ui/native-select'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { TitledCard } from '@/components/ui/titled-card'
import { cn } from '@/lib/utils'
import {
  cancelPrivacyRequest,
  createPrivacyRequest,
  getPersonalInfoSnapshot,
  listMyPrivacyRequests,
} from '../api'
import type {
  CreatePrivacyRequestPayload,
  PersonalInfoSnapshot,
  PrivacyRequest,
  PrivacyRequestStatus,
  PrivacyRequestType,
} from '../types'
import { formatComplianceTimestamp, normalizePagePayload } from '../utils'

const DEFAULT_PRIVACY_FORM: CreatePrivacyRequestPayload = {
  request_type: 'access',
  contact_name: '',
  contact_email: '',
  contact_phone: '',
  content: '',
}

function getPrivacyTypeLabel(
  t: (key: string) => string,
  value: PrivacyRequestType | string
) {
  const labels: Record<string, string> = {
    access: t('Access and copy'),
    correction: t('Correction'),
    deletion: t('Deletion'),
  }
  return labels[value] ?? value
}

function getPrivacyStatusLabel(
  t: (key: string) => string,
  value: PrivacyRequestStatus | string
) {
  const labels: Record<string, string> = {
    pending: t('Pending'),
    processing: t('Processing'),
    completed: t('Completed'),
    rejected: t('Rejected'),
    cancelled: t('Cancelled'),
  }
  return labels[value] ?? value
}

function getStatusBadgeClass(status: PrivacyRequestStatus | string) {
  if (status === 'pending') return 'border-amber-200 bg-amber-50 text-amber-700'
  if (status === 'processing') return 'border-sky-200 bg-sky-50 text-sky-700'
  if (status === 'completed') return 'border-emerald-200 bg-emerald-50 text-emerald-700'
  if (status === 'rejected') return 'border-rose-200 bg-rose-50 text-rose-700'
  return 'border-muted bg-muted text-muted-foreground'
}

function formatSnapshotValue(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  if (typeof value === 'number') return String(value)
  return String(value)
}

function buildSnapshotRows(
  t: (key: string) => string,
  snapshot?: PersonalInfoSnapshot
) {
  if (!snapshot) return []
  return [
    [t('User ID'), snapshot.id],
    [t('Username'), snapshot.username],
    [t('Display name'), snapshot.display_name],
    [t('Email'), snapshot.email],
    [t('Phone number'), snapshot.phone_number],
    [t('Role'), snapshot.role],
    [t('Status'), snapshot.status],
    [t('Group'), snapshot.group],
    [t('Quota'), snapshot.quota],
    [t('Used quota'), snapshot.used_quota],
    [t('Request count'), snapshot.request_count],
    [t('Inviter ID'), snapshot.inviter_id],
    [t('Created at'), formatComplianceTimestamp(snapshot.created_at)],
    [t('Last login at'), formatComplianceTimestamp(snapshot.last_login_at)],
    [t('GitHub ID'), snapshot.github_id],
    [t('Discord ID'), snapshot.discord_id],
    [t('OIDC ID'), snapshot.oidc_id],
    [t('WeChat ID'), snapshot.wechat_id],
    [t('Telegram ID'), snapshot.telegram_id],
    [t('Linux DO ID'), snapshot.linux_do_id],
  ]
}

function canCancelRequest(record: PrivacyRequest) {
  return record.status === 'pending' || record.status === 'processing'
}

export function PersonalInfoRightsCard() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [form, setForm] =
    useState<CreatePrivacyRequestPayload>(DEFAULT_PRIVACY_FORM)
  const [snapshotPrefilled, setSnapshotPrefilled] = useState(false)

  const snapshotQuery = useQuery({
    queryKey: ['privacy-personal-info'],
    queryFn: getPersonalInfoSnapshot,
  })

  const requestsQuery = useQuery({
    queryKey: ['privacy-requests', 'me'],
    queryFn: () => listMyPrivacyRequests({ p: 1, page_size: 20 }),
  })

  const snapshot = snapshotQuery.data?.data
  const requests = normalizePagePayload(requestsQuery.data?.data).items
  const snapshotRows = buildSnapshotRows(t, snapshot)

  useEffect(() => {
    if (!snapshot || snapshotPrefilled) return
    setForm((previous) => ({
      ...previous,
      contact_name:
        previous.contact_name || snapshot.display_name || snapshot.username || '',
      contact_email: previous.contact_email || snapshot.email || '',
      contact_phone: previous.contact_phone || snapshot.phone_number || '',
    }))
    setSnapshotPrefilled(true)
  }, [snapshot, snapshotPrefilled])

  const createRequestMutation = useMutation({
    mutationFn: createPrivacyRequest,
    onSuccess: (response) => {
      if (response.success === false) {
        toast.error(response.message || t('Failed to submit request'))
        return
      }
      toast.success(t('Request submitted successfully'))
      setForm((previous) => ({
        ...DEFAULT_PRIVACY_FORM,
        contact_name: previous.contact_name,
        contact_email: previous.contact_email,
        contact_phone: previous.contact_phone,
      }))
      queryClient.invalidateQueries({ queryKey: ['privacy-requests', 'me'] })
    },
  })

  const cancelRequestMutation = useMutation({
    mutationFn: cancelPrivacyRequest,
    onSuccess: (response) => {
      if (response.success === false) {
        toast.error(response.message || t('Failed to cancel request'))
        return
      }
      toast.success(t('Request cancelled successfully'))
      queryClient.invalidateQueries({ queryKey: ['privacy-requests', 'me'] })
    },
  })

  const updateFormField = (
    field: keyof CreatePrivacyRequestPayload,
    value: string
  ) => {
    setForm((previous) => ({ ...previous, [field]: value }))
  }

  const submitRequest = () => {
    const payload: CreatePrivacyRequestPayload = {
      request_type: form.request_type,
      contact_name: form.contact_name.trim(),
      contact_email: form.contact_email.trim(),
      contact_phone: form.contact_phone.trim(),
      content: form.content.trim(),
    }
    if (!payload.content) {
      toast.error(t('Please describe your request'))
      return
    }
    createRequestMutation.mutate(payload)
  }

  const cancelRequest = (record: PrivacyRequest) => {
    const confirmed = window.confirm(
      t('Cancel this personal information rights request?')
    )
    if (!confirmed) return
    cancelRequestMutation.mutate(record.id)
  }

  return (
    <TitledCard
      title={t('Personal information rights')}
      description={t(
        'View your personal information snapshot and submit access, correction, or deletion requests.'
      )}
      icon={<ShieldCheck className='size-4' />}
    >
      <div className='space-y-5'>
        <div className='rounded-lg border'>
          <div className='border-b px-3 py-2 text-sm font-medium'>
            {t('Personal information snapshot')}
          </div>
          {snapshotQuery.isLoading ? (
            <div className='grid gap-2 p-3 sm:grid-cols-2'>
              {Array.from({ length: 8 }).map((_, index) => (
                <Skeleton key={index} className='h-8 w-full' />
              ))}
            </div>
          ) : (
            <dl className='grid gap-px overflow-hidden rounded-b-lg bg-border sm:grid-cols-2'>
              {snapshotRows.map(([label, value]) => (
                <div key={String(label)} className='bg-card px-3 py-2'>
                  <dt className='text-muted-foreground text-xs'>{label}</dt>
                  <dd className='break-all text-sm'>
                    {formatSnapshotValue(value)}
                  </dd>
                </div>
              ))}
            </dl>
          )}
        </div>

        <div className='grid gap-3 sm:grid-cols-2'>
          <div className='space-y-1.5'>
            <Label htmlFor='privacy-request-type'>{t('Request type')}</Label>
            <NativeSelect
              id='privacy-request-type'
              className='w-full'
              value={form.request_type}
              onChange={(event) =>
                updateFormField(
                  'request_type',
                  event.target.value as PrivacyRequestType
                )
              }
            >
              <NativeSelectOption value='access'>
                {t('Access and copy')}
              </NativeSelectOption>
              <NativeSelectOption value='correction'>
                {t('Correction')}
              </NativeSelectOption>
              <NativeSelectOption value='deletion'>
                {t('Deletion')}
              </NativeSelectOption>
            </NativeSelect>
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='privacy-contact-name'>{t('Contact name')}</Label>
            <Input
              id='privacy-contact-name'
              value={form.contact_name}
              onChange={(event) =>
                updateFormField('contact_name', event.target.value)
              }
              placeholder={t('Enter contact name')}
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='privacy-contact-email'>{t('Email')}</Label>
            <Input
              id='privacy-contact-email'
              type='email'
              value={form.contact_email}
              onChange={(event) =>
                updateFormField('contact_email', event.target.value)
              }
              placeholder={t('Enter email')}
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='privacy-contact-phone'>{t('Phone number')}</Label>
            <Input
              id='privacy-contact-phone'
              value={form.contact_phone}
              onChange={(event) =>
                updateFormField('contact_phone', event.target.value)
              }
              placeholder={t('Enter phone number')}
            />
          </div>
          <div className='space-y-1.5 sm:col-span-2'>
            <Label htmlFor='privacy-request-content'>
              {t('Request details')}
            </Label>
            <Textarea
              id='privacy-request-content'
              value={form.content}
              onChange={(event) =>
                updateFormField('content', event.target.value)
              }
              placeholder={t(
                'Describe the information you want to access, correct, or delete.'
              )}
              className='min-h-24'
            />
          </div>
        </div>

        <div className='flex justify-end'>
          <Button
            onClick={submitRequest}
            disabled={createRequestMutation.isPending}
          >
            <FileText className='size-4' />
            {createRequestMutation.isPending
              ? t('Submitting...')
              : t('Submit request')}
          </Button>
        </div>

        <div className='space-y-2'>
          <h3 className='text-sm font-medium'>{t('My requests')}</h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Type')}</TableHead>
                <TableHead>{t('Status')}</TableHead>
                <TableHead>{t('Submitted at')}</TableHead>
                <TableHead>{t('Handled at')}</TableHead>
                <TableHead>{t('Admin note')}</TableHead>
                <TableHead className='text-right'>{t('Actions')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {requestsQuery.isLoading ? (
                <TableRow>
                  <TableCell colSpan={6}>
                    <Skeleton className='h-10 w-full' />
                  </TableCell>
                </TableRow>
              ) : requests.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className='text-muted-foreground h-16 text-center'
                  >
                    {t('No requests yet')}
                  </TableCell>
                </TableRow>
              ) : (
                requests.map((record) => (
                  <TableRow key={record.id}>
                    <TableCell>
                      {getPrivacyTypeLabel(t, record.request_type)}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant='outline'
                        className={cn(getStatusBadgeClass(record.status))}
                      >
                        {getPrivacyStatusLabel(t, record.status)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {formatComplianceTimestamp(record.created_at)}
                    </TableCell>
                    <TableCell>
                      {formatComplianceTimestamp(record.handled_at)}
                    </TableCell>
                    <TableCell className='max-w-56 whitespace-normal'>
                      {record.admin_note || '-'}
                    </TableCell>
                    <TableCell className='text-right'>
                      {canCancelRequest(record) ? (
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => cancelRequest(record)}
                          disabled={cancelRequestMutation.isPending}
                        >
                          {t('Cancel')}
                        </Button>
                      ) : (
                        <span className='text-muted-foreground text-xs'>-</span>
                      )}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </TitledCard>
  )
}
