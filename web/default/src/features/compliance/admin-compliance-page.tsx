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
import { useEffect, useState, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ClipboardList, Eye, MessageSquare, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { Dialog } from '@/components/dialog'
import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import {
  NativeSelect,
  NativeSelectOption,
} from '@/components/ui/native-select'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import {
  ADMIN_PERMISSION_ACTIONS,
  ADMIN_PERMISSION_RESOURCES,
  hasPermission,
} from '@/lib/admin-permissions'
import { cn } from '@/lib/utils'
import {
  getAdminFeedback,
  getAdminPrivacyRequest,
  listAdminFeedback,
  listAdminPrivacyRequests,
  updateAdminFeedback,
  updateAdminPrivacyRequest,
} from './api'
import type {
  PrivacyRequest,
  PrivacyRequestStatus,
  PrivacyRequestType,
  PublicFeedback,
  PublicFeedbackStatus,
  PublicFeedbackType,
} from './types'
import { formatComplianceTimestamp, normalizePagePayload } from './utils'

type ComplianceTab = 'privacy' | 'feedback'

type DetailState = {
  kind: ComplianceTab
  id: number
}

type DetailFieldProps = {
  label: ReactNode
  value: ReactNode
}

type PaginationControlsProps = {
  page: number
  pageSize: number
  total: number
  onPageChange: (page: number) => void
  onPageSizeChange: (pageSize: number) => void
}

type Filters = {
  status: string
  type: string
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

function getFeedbackTypeLabel(
  t: (key: string) => string,
  value: PublicFeedbackType | string
) {
  const labels: Record<string, string> = {
    complaint: t('Complaint'),
    feedback: t('Feedback'),
    other: t('Other'),
  }
  return labels[value] ?? value
}

function getStatusLabel(t: (key: string) => string, value: string) {
  const labels: Record<string, string> = {
    pending: t('Pending'),
    processing: t('Processing'),
    completed: t('Completed'),
    resolved: t('Resolved'),
    closed: t('Closed'),
    rejected: t('Rejected'),
    cancelled: t('Cancelled'),
  }
  return labels[value] ?? value
}

function getStatusBadgeClass(status: string) {
  if (status === 'pending') return 'border-amber-200 bg-amber-50 text-amber-700'
  if (status === 'processing') return 'border-sky-200 bg-sky-50 text-sky-700'
  if (status === 'completed' || status === 'resolved')
    return 'border-emerald-200 bg-emerald-50 text-emerald-700'
  if (status === 'rejected') return 'border-rose-200 bg-rose-50 text-rose-700'
  return 'border-muted bg-muted text-muted-foreground'
}

function getDisplayValue(value: ReactNode) {
  if (value === null || value === undefined || value === '') return '-'
  return value
}

function DetailField(props: DetailFieldProps) {
  return (
    <div className='border-b border-dashed py-2 last:border-b-0'>
      <div className='text-muted-foreground mb-1 text-xs'>{props.label}</div>
      <div className='break-all whitespace-pre-wrap text-sm'>
        {getDisplayValue(props.value)}
      </div>
    </div>
  )
}

function StatusBadge(props: { status: string; label: string }) {
  return (
    <Badge
      variant='outline'
      className={cn(getStatusBadgeClass(props.status))}
    >
      {props.label}
    </Badge>
  )
}

function PaginationControls(props: PaginationControlsProps) {
  const { t } = useTranslation()
  const totalPages = Math.max(1, Math.ceil(props.total / props.pageSize))

  return (
    <div className='flex flex-col gap-2 border-t px-3 py-2 sm:flex-row sm:items-center sm:justify-between'>
      <div className='text-muted-foreground text-sm'>
        {t('Total')}: {props.total}
      </div>
      <div className='flex flex-wrap items-center gap-2'>
        <NativeSelect
          size='sm'
          value={String(props.pageSize)}
          onChange={(event) =>
            props.onPageSizeChange(Number(event.target.value))
          }
        >
          {[10, 20, 50, 100].map((size) => (
            <NativeSelectOption key={size} value={String(size)}>
              {size} / {t('page')}
            </NativeSelectOption>
          ))}
        </NativeSelect>
        <Button
          variant='outline'
          size='sm'
          disabled={props.page <= 1}
          onClick={() => props.onPageChange(Math.max(1, props.page - 1))}
        >
          {t('Previous')}
        </Button>
        <span className='text-muted-foreground min-w-20 text-center text-sm'>
          {props.page} / {totalPages}
        </span>
        <Button
          variant='outline'
          size='sm'
          disabled={props.page >= totalPages}
          onClick={() => props.onPageChange(Math.min(totalPages, props.page + 1))}
        >
          {t('Next')}
        </Button>
      </div>
    </div>
  )
}

function isPrivacyRecord(
  record: PrivacyRequest | PublicFeedback | null
): record is PrivacyRequest {
  return Boolean(record && 'request_type' in record)
}

function isFeedbackRecord(
  record: PrivacyRequest | PublicFeedback | null
): record is PublicFeedback {
  return Boolean(record && 'feedback_type' in record)
}

export function AdminCompliancePage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const user = useAuthStore((state) => state.auth.user)
  const canWrite = hasPermission(
    user,
    ADMIN_PERMISSION_RESOURCES.COMPLIANCE,
    ADMIN_PERMISSION_ACTIONS.WRITE
  )
  const [activeTab, setActiveTab] = useState<ComplianceTab>('privacy')
  const [filters, setFilters] = useState<Filters>({ status: '', type: '' })
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [detail, setDetail] = useState<DetailState | null>(null)
  const [fallbackRecord, setFallbackRecord] = useState<
    PrivacyRequest | PublicFeedback | null
  >(null)
  const [nextStatus, setNextStatus] = useState('')
  const [adminNote, setAdminNote] = useState('')
  const [executeAccountDeletion, setExecuteAccountDeletion] = useState(false)

  const privacyQuery = useQuery({
    queryKey: ['admin-privacy-requests', page, pageSize, filters],
    queryFn: () =>
      listAdminPrivacyRequests({
        p: page,
        page_size: pageSize,
        status: filters.status || undefined,
        request_type: filters.type || undefined,
      }),
    enabled: activeTab === 'privacy',
  })

  const feedbackQuery = useQuery({
    queryKey: ['admin-feedback', page, pageSize, filters],
    queryFn: () =>
      listAdminFeedback({
        p: page,
        page_size: pageSize,
        status: filters.status || undefined,
        feedback_type: filters.type || undefined,
      }),
    enabled: activeTab === 'feedback',
  })

  const privacyDetailQuery = useQuery({
    queryKey: ['admin-privacy-request', detail?.id],
    queryFn: () => getAdminPrivacyRequest(detail?.id ?? 0),
    enabled: detail?.kind === 'privacy' && Boolean(detail?.id),
  })

  const feedbackDetailQuery = useQuery({
    queryKey: ['admin-feedback-detail', detail?.id],
    queryFn: () => getAdminFeedback(detail?.id ?? 0),
    enabled: detail?.kind === 'feedback' && Boolean(detail?.id),
  })

  const updatePrivacyMutation = useMutation({
    mutationFn: (payload: {
      id: number
      status: PrivacyRequestStatus
      adminNote: string
      executeAccountDeletion: boolean
    }) =>
      updateAdminPrivacyRequest(payload.id, {
        status: payload.status,
        admin_note: payload.adminNote,
        execute_account_deletion: payload.executeAccountDeletion,
      }),
    onSuccess: (response) => {
      if (response.success === false) {
        toast.error(response.message || t('Failed to update status'))
        return
      }
      toast.success(t('Status updated successfully'))
      setDetail(null)
      setFallbackRecord(null)
      queryClient.invalidateQueries({ queryKey: ['admin-privacy-requests'] })
      queryClient.invalidateQueries({ queryKey: ['admin-privacy-request'] })
    },
  })

  const updateFeedbackMutation = useMutation({
    mutationFn: (payload: {
      id: number
      status: PublicFeedbackStatus
      adminNote: string
    }) =>
      updateAdminFeedback(payload.id, {
        status: payload.status,
        admin_note: payload.adminNote,
      }),
    onSuccess: (response) => {
      if (response.success === false) {
        toast.error(response.message || t('Failed to update status'))
        return
      }
      toast.success(t('Status updated successfully'))
      setDetail(null)
      setFallbackRecord(null)
      queryClient.invalidateQueries({ queryKey: ['admin-feedback'] })
      queryClient.invalidateQueries({ queryKey: ['admin-feedback-detail'] })
    },
  })

  const privacyPage = normalizePagePayload(privacyQuery.data?.data)
  const feedbackPage = normalizePagePayload(feedbackQuery.data?.data)
  const activeLoading =
    activeTab === 'privacy' ? privacyQuery.isLoading : feedbackQuery.isLoading
  const detailRecord = !detail
    ? null
    : detail.kind === 'privacy'
      ? privacyDetailQuery.data?.data ?? fallbackRecord
      : feedbackDetailQuery.data?.data ?? fallbackRecord
  const isDetailPrivacy = detail?.kind === 'privacy'
  const isUpdating =
    updatePrivacyMutation.isPending || updateFeedbackMutation.isPending

  useEffect(() => {
    if (!detailRecord) return
    setNextStatus(detailRecord.status)
    setAdminNote(detailRecord.admin_note || '')
    setExecuteAccountDeletion(
      isPrivacyRecord(detailRecord)
        ? Boolean(detailRecord.execute_account_deletion)
        : false
    )
  }, [detailRecord])

  const changeTab = (value: string | null) => {
    if (value !== 'privacy' && value !== 'feedback') return
    setActiveTab(value)
    setFilters({ status: '', type: '' })
    setPage(1)
  }

  const updateFilter = (field: keyof Filters, value: string) => {
    setFilters((previous) => ({ ...previous, [field]: value }))
    setPage(1)
  }

  const updatePageSize = (value: number) => {
    setPageSize(value)
    setPage(1)
  }

  const openPrivacyDetail = (record: PrivacyRequest) => {
    setFallbackRecord(record)
    setDetail({ kind: 'privacy', id: record.id })
  }

  const openFeedbackDetail = (record: PublicFeedback) => {
    setFallbackRecord(record)
    setDetail({ kind: 'feedback', id: record.id })
  }

  const closeDetail = () => {
    setDetail(null)
    setFallbackRecord(null)
  }

  const submitDetailUpdate = () => {
    if (!canWrite || !detail || !detailRecord) return
    if (detail.kind === 'privacy' && isPrivacyRecord(detailRecord)) {
      updatePrivacyMutation.mutate({
        id: detailRecord.id,
        status: nextStatus as PrivacyRequestStatus,
        adminNote: adminNote.trim(),
        executeAccountDeletion:
          detailRecord.request_type === 'deletion' && executeAccountDeletion,
      })
      return
    }
    if (detail.kind === 'feedback' && isFeedbackRecord(detailRecord)) {
      updateFeedbackMutation.mutate({
        id: detailRecord.id,
        status: nextStatus as PublicFeedbackStatus,
        adminNote: adminNote.trim(),
      })
    }
  }

  const privacyStatusOptions = [
    ['', t('All statuses')],
    ['pending', t('Pending')],
    ['processing', t('Processing')],
    ['completed', t('Completed')],
    ['rejected', t('Rejected')],
    ['cancelled', t('Cancelled')],
  ]
  const feedbackStatusOptions = [
    ['', t('All statuses')],
    ['pending', t('Pending')],
    ['processing', t('Processing')],
    ['resolved', t('Resolved')],
    ['closed', t('Closed')],
    ['rejected', t('Rejected')],
  ]
  const privacyTypeOptions = [
    ['', t('All types')],
    ['access', t('Access and copy')],
    ['correction', t('Correction')],
    ['deletion', t('Deletion')],
  ]
  const feedbackTypeOptions = [
    ['', t('All types')],
    ['complaint', t('Complaint')],
    ['feedback', t('Feedback')],
    ['other', t('Other')],
  ]
  const detailKind = detail?.kind ?? activeTab
  const updateStatusOptions =
    detailKind === 'privacy'
      ? privacyStatusOptions.filter(([value]) => value)
      : feedbackStatusOptions.filter(([value]) => value)
  const filterStatusOptions =
    activeTab === 'privacy' ? privacyStatusOptions : feedbackStatusOptions
  const filterTypeOptions =
    activeTab === 'privacy' ? privacyTypeOptions : feedbackTypeOptions
  const showDeletionSwitch =
    isDetailPrivacy &&
    isPrivacyRecord(detailRecord) &&
    detailRecord.request_type === 'deletion'

  return (
    <>
      <SectionPageLayout fixedContent>
        <SectionPageLayout.Title>
          <span className='flex items-center gap-2'>
            <ShieldCheck className='size-4' />
            {t('Compliance')}
          </span>
        </SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <Card className='flex h-full min-h-0'>
            <CardContent className='flex min-h-0 flex-1 flex-col p-3 sm:p-4'>
              <Tabs
                value={activeTab}
                onValueChange={changeTab}
                className='min-h-0 flex-1'
              >
                <div className='flex flex-col gap-3 border-b pb-3 lg:flex-row lg:items-center lg:justify-between'>
                  <TabsList>
                    <TabsTrigger value='privacy'>
                      <ClipboardList className='size-4' />
                      {t('Personal information requests')}
                    </TabsTrigger>
                    <TabsTrigger value='feedback'>
                      <MessageSquare className='size-4' />
                      {t('Public feedback')}
                    </TabsTrigger>
                  </TabsList>
                  <div className='flex flex-wrap gap-2'>
                    <NativeSelect
                      value={filters.status}
                      onChange={(event) =>
                        updateFilter('status', event.target.value)
                      }
                    >
                      {filterStatusOptions.map(([value, label]) => (
                        <NativeSelectOption key={value || 'all'} value={value}>
                          {label}
                        </NativeSelectOption>
                      ))}
                    </NativeSelect>
                    <NativeSelect
                      value={filters.type}
                      onChange={(event) =>
                        updateFilter('type', event.target.value)
                      }
                    >
                      {filterTypeOptions.map(([value, label]) => (
                        <NativeSelectOption key={value || 'all'} value={value}>
                          {label}
                        </NativeSelectOption>
                      ))}
                    </NativeSelect>
                  </div>
                </div>

                <TabsContent value='privacy' className='min-h-0 flex-1 pt-3'>
                  <div className='flex h-full min-h-0 flex-col rounded-lg border'>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>{t('ID')}</TableHead>
                          <TableHead>{t('Applicant')}</TableHead>
                          <TableHead>{t('Type')}</TableHead>
                          <TableHead>{t('Status')}</TableHead>
                          <TableHead>{t('Contact')}</TableHead>
                          <TableHead>{t('Submitted at')}</TableHead>
                          <TableHead className='text-right'>
                            {t('Actions')}
                          </TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {activeLoading ? (
                          <TableRow>
                            <TableCell colSpan={7}>
                              <Skeleton className='h-10 w-full' />
                            </TableCell>
                          </TableRow>
                        ) : privacyPage.items.length === 0 ? (
                          <TableRow>
                            <TableCell
                              colSpan={7}
                              className='text-muted-foreground h-24 text-center'
                            >
                              {t('No records found')}
                            </TableCell>
                          </TableRow>
                        ) : (
                          privacyPage.items.map((record) => (
                            <TableRow key={record.id}>
                              <TableCell>{record.id}</TableCell>
                              <TableCell>
                                {record.username || record.user_id || '-'}
                              </TableCell>
                              <TableCell>
                                {getPrivacyTypeLabel(t, record.request_type)}
                              </TableCell>
                              <TableCell>
                                <StatusBadge
                                  status={record.status}
                                  label={getStatusLabel(t, record.status)}
                                />
                              </TableCell>
                              <TableCell className='max-w-52 whitespace-normal'>
                                {record.contact_name || record.contact_email ||
                                  record.contact_phone ||
                                  '-'}
                              </TableCell>
                              <TableCell>
                                {formatComplianceTimestamp(record.created_at)}
                              </TableCell>
                              <TableCell className='text-right'>
                                <Button
                                  variant='outline'
                                  size='sm'
                                  onClick={() => openPrivacyDetail(record)}
                                >
                                  <Eye className='size-4' />
                                  {t('View')}
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))
                        )}
                      </TableBody>
                    </Table>
                    <PaginationControls
                      page={page}
                      pageSize={pageSize}
                      total={privacyPage.total}
                      onPageChange={setPage}
                      onPageSizeChange={updatePageSize}
                    />
                  </div>
                </TabsContent>

                <TabsContent value='feedback' className='min-h-0 flex-1 pt-3'>
                  <div className='flex h-full min-h-0 flex-col rounded-lg border'>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>{t('Tracking code')}</TableHead>
                          <TableHead>{t('Contact')}</TableHead>
                          <TableHead>{t('Type')}</TableHead>
                          <TableHead>{t('Status')}</TableHead>
                          <TableHead>{t('Title')}</TableHead>
                          <TableHead>{t('Submitted at')}</TableHead>
                          <TableHead className='text-right'>
                            {t('Actions')}
                          </TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {activeLoading ? (
                          <TableRow>
                            <TableCell colSpan={7}>
                              <Skeleton className='h-10 w-full' />
                            </TableCell>
                          </TableRow>
                        ) : feedbackPage.items.length === 0 ? (
                          <TableRow>
                            <TableCell
                              colSpan={7}
                              className='text-muted-foreground h-24 text-center'
                            >
                              {t('No records found')}
                            </TableCell>
                          </TableRow>
                        ) : (
                          feedbackPage.items.map((record) => (
                            <TableRow key={record.id}>
                              <TableCell>{record.tracking_code}</TableCell>
                              <TableCell>
                                {record.contact_name ||
                                  record.username ||
                                  record.user_id ||
                                  '-'}
                              </TableCell>
                              <TableCell>
                                {getFeedbackTypeLabel(t, record.feedback_type)}
                              </TableCell>
                              <TableCell>
                                <StatusBadge
                                  status={record.status}
                                  label={getStatusLabel(t, record.status)}
                                />
                              </TableCell>
                              <TableCell className='max-w-72 whitespace-normal'>
                                {record.title}
                              </TableCell>
                              <TableCell>
                                {formatComplianceTimestamp(record.created_at)}
                              </TableCell>
                              <TableCell className='text-right'>
                                <Button
                                  variant='outline'
                                  size='sm'
                                  onClick={() => openFeedbackDetail(record)}
                                >
                                  <Eye className='size-4' />
                                  {t('View')}
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))
                        )}
                      </TableBody>
                    </Table>
                    <PaginationControls
                      page={page}
                      pageSize={pageSize}
                      total={feedbackPage.total}
                      onPageChange={setPage}
                      onPageSizeChange={updatePageSize}
                    />
                  </div>
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <Dialog
        open={Boolean(detail)}
        onOpenChange={(open) => {
          if (!open) closeDetail()
        }}
        title={
          isDetailPrivacy
            ? t('Personal information request details')
            : t('Public feedback details')
        }
        description={t('Review the record and update the processing status.')}
        contentClassName='sm:max-w-3xl'
        footer={
          <>
            <Button variant='outline' onClick={closeDetail}>
              {t('Close')}
            </Button>
            <Button
              onClick={submitDetailUpdate}
              disabled={!canWrite || isUpdating || !detailRecord}
            >
              {isUpdating ? t('Saving...') : t('Save')}
            </Button>
          </>
        }
      >
        <div className='space-y-4'>
          {!canWrite ? (
            <Alert>
              <AlertTitle>{t('Read-only access')}</AlertTitle>
              <AlertDescription>
                {t('You do not have permission to update compliance records.')}
              </AlertDescription>
            </Alert>
          ) : null}

          {!detailRecord ? (
            <Skeleton className='h-48 w-full' />
          ) : (
            <>
              <div className='grid gap-3 sm:grid-cols-2'>
                {isPrivacyRecord(detailRecord) ? (
                  <>
                    <DetailField label={t('ID')} value={detailRecord.id} />
                    <DetailField
                      label={t('Applicant')}
                      value={detailRecord.username || detailRecord.user_id}
                    />
                    <DetailField
                      label={t('Type')}
                      value={getPrivacyTypeLabel(t, detailRecord.request_type)}
                    />
                    <DetailField
                      label={t('Submitted at')}
                      value={formatComplianceTimestamp(detailRecord.created_at)}
                    />
                    <DetailField
                      label={t('Contact name')}
                      value={detailRecord.contact_name}
                    />
                    <DetailField
                      label={t('Email')}
                      value={detailRecord.contact_email}
                    />
                    <DetailField
                      label={t('Phone number')}
                      value={detailRecord.contact_phone}
                    />
                    <DetailField
                      label={t('Handled at')}
                      value={formatComplianceTimestamp(detailRecord.handled_at)}
                    />
                    <div className='sm:col-span-2'>
                      <DetailField
                        label={t('Request details')}
                        value={detailRecord.content}
                      />
                    </div>
                  </>
                ) : null}

                {isFeedbackRecord(detailRecord) ? (
                  <>
                    <DetailField
                      label={t('Tracking code')}
                      value={detailRecord.tracking_code}
                    />
                    <DetailField
                      label={t('Linked user')}
                      value={detailRecord.username || detailRecord.user_id || '-'}
                    />
                    <DetailField
                      label={t('Type')}
                      value={getFeedbackTypeLabel(t, detailRecord.feedback_type)}
                    />
                    <DetailField
                      label={t('Submitted at')}
                      value={formatComplianceTimestamp(detailRecord.created_at)}
                    />
                    <DetailField
                      label={t('Contact name')}
                      value={detailRecord.contact_name}
                    />
                    <DetailField
                      label={t('Email')}
                      value={detailRecord.contact_email}
                    />
                    <DetailField
                      label={t('Phone number')}
                      value={detailRecord.contact_phone}
                    />
                    <DetailField label={t('Title')} value={detailRecord.title} />
                    <div className='sm:col-span-2'>
                      <DetailField
                        label={t('Content')}
                        value={detailRecord.content}
                      />
                    </div>
                  </>
                ) : null}
              </div>

              <div className='grid gap-3 sm:grid-cols-2'>
                <div className='space-y-1.5'>
                  <Label htmlFor='compliance-next-status'>
                    {t('Processing status')}
                  </Label>
                  <NativeSelect
                    id='compliance-next-status'
                    className='w-full'
                    value={nextStatus}
                    disabled={!canWrite}
                    onChange={(event) => setNextStatus(event.target.value)}
                  >
                    {updateStatusOptions.map(([value, label]) => (
                      <NativeSelectOption key={value} value={value}>
                        {label}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>
                </div>
                {showDeletionSwitch ? (
                  <div className='flex items-center justify-between gap-3 rounded-lg border px-3 py-2'>
                    <div className='space-y-1'>
                      <Label>{t('Execute account closure')}</Label>
                      <p className='text-muted-foreground text-xs'>
                        {t(
                          'Only enable this after the deletion request is completed.'
                        )}
                      </p>
                    </div>
                    <Switch
                      checked={executeAccountDeletion}
                      disabled={!canWrite}
                      onCheckedChange={setExecuteAccountDeletion}
                    />
                  </div>
                ) : null}
                <div className='space-y-1.5 sm:col-span-2'>
                  <Label htmlFor='compliance-admin-note'>
                    {t('Processing note')}
                  </Label>
                  <Textarea
                    id='compliance-admin-note'
                    value={adminNote}
                    disabled={!canWrite}
                    onChange={(event) => setAdminNote(event.target.value)}
                    placeholder={t('Enter processing notes')}
                    className='min-h-28'
                  />
                </div>
              </div>
            </>
          )}
        </div>
      </Dialog>
    </>
  )
}
