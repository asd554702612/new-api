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
import { useQuery } from '@tanstack/react-query'
import { RefreshCw, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { Button } from '@/components/ui/button'
import { Empty, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Input } from '@/components/ui/input'
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  sideDrawerContentClassName,
  sideDrawerFormClassName,
  sideDrawerHeaderClassName,
} from '@/components/drawer-layout'
import { GroupBadge } from '@/components/group-badge'
import { StatusBadge } from '@/components/status-badge'
import { TableId } from '@/components/table-id'
import { getPlanSubscriptions } from '../../api'
import { formatTimestamp } from '../../lib'
import type {
  PlanRecord,
  PlanSubscriptionRecord,
  PlanSubscriptionStatusFilter,
  UserSubscription,
} from '../../types'

const PAGE_SIZE = 10

interface PlanSubscriptionsDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: PlanRecord | null
}

function getUserDisplayName(record: PlanSubscriptionRecord): string {
  return (
    record.user.display_name ||
    record.user.username ||
    record.user.email ||
    `#${record.user.id}`
  )
}

function SubscriptionStatusBadge(props: { subscription: UserSubscription }) {
  const { t } = useTranslation()
  const now = Date.now() / 1000
  const isExpired =
    (props.subscription.end_time || 0) > 0 && props.subscription.end_time < now
  const isActive = props.subscription.status === 'active' && !isExpired

  if (isActive) {
    return (
      <StatusBadge label={t('Active')} variant='success' copyable={false} />
    )
  }
  if (props.subscription.status === 'cancelled') {
    return (
      <StatusBadge label={t('Cancelled')} variant='neutral' copyable={false} />
    )
  }
  return <StatusBadge label={t('Expired')} variant='neutral' copyable={false} />
}

function UsageText(props: { subscription: UserSubscription }) {
  const { t } = useTranslation()
  const total = Number(props.subscription.amount_total || 0)
  const used = Number(props.subscription.amount_used || 0)
  if (total <= 0) return <>{t('Unlimited')}</>
  return (
    <>
      {formatQuota(used)} / {formatQuota(total)}
    </>
  )
}

function BuyerSummary(props: { record: PlanSubscriptionRecord }) {
  const { t } = useTranslation()
  const user = props.record.user

  return (
    <div className='min-w-0'>
      <div className='flex min-w-0 items-center gap-2'>
        <span className='truncate font-medium'>
          {getUserDisplayName(props.record)}
        </span>
        <TableId value={user.id} />
      </div>
      <div className='text-muted-foreground mt-1 flex min-w-0 flex-wrap gap-x-3 gap-y-1 text-xs'>
        <span className='truncate'>{user.email || '-'}</span>
        {user.group && <GroupBadge group={user.group} />}
        <span>
          {t('Created At')}: {formatTimestamp(user.created_at)}
        </span>
      </div>
    </div>
  )
}

function MobileSubscriptionCard(props: { record: PlanSubscriptionRecord }) {
  const { t } = useTranslation()
  const sub = props.record.subscription

  return (
    <div className='border-border/70 flex flex-col gap-3 border-b py-3 last:border-b-0'>
      <div className='flex items-start justify-between gap-3'>
        <BuyerSummary record={props.record} />
        <SubscriptionStatusBadge subscription={sub} />
      </div>
      <div className='grid grid-cols-2 gap-3 text-xs'>
        <div>
          <div className='text-muted-foreground'>{t('Subscription')}</div>
          <TableId value={sub.id} />
        </div>
        <div>
          <div className='text-muted-foreground'>{t('Usage')}</div>
          <UsageText subscription={sub} />
        </div>
        <div>
          <div className='text-muted-foreground'>{t('Validity')}</div>
          <div>{formatTimestamp(sub.start_time)}</div>
          <div>{formatTimestamp(sub.end_time)}</div>
        </div>
        <div>
          <div className='text-muted-foreground'>{t('Source')}</div>
          <div>{sub.source || '-'}</div>
        </div>
      </div>
    </div>
  )
}

export function PlanSubscriptionsDrawer(props: PlanSubscriptionsDrawerProps) {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [keywordInput, setKeywordInput] = useState('')
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<PlanSubscriptionStatusFilter>('all')

  const plan = props.currentRow?.plan
  const planId = plan?.id || 0

  useEffect(() => {
    if (props.open) {
      setPage(1)
      setKeywordInput('')
      setKeyword('')
      setStatus('all')
    }
  }, [props.open, planId])

  const query = useQuery({
    queryKey: [
      'admin-plan-subscriptions',
      planId,
      page,
      PAGE_SIZE,
      status,
      keyword,
    ],
    enabled: props.open && planId > 0,
    queryFn: async () => {
      const result = await getPlanSubscriptions({
        planId,
        page,
        pageSize: PAGE_SIZE,
        status,
        keyword,
      })
      if (!result.success || !result.data) {
        throw new Error(result.message || 'Request failed')
      }
      return result.data
    },
    placeholderData: (prev) => prev,
  })

  const records = useMemo(() => query.data?.items || [], [query.data?.items])
  const total = query.data?.total || 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const submitSearch = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setPage(1)
    setKeyword(keywordInput.trim())
  }

  const changeStatus = (value: string | null) => {
    setPage(1)
    setStatus((value || 'all') as PlanSubscriptionStatusFilter)
  }

  const goToPage = (nextPage: number) => {
    setPage(Math.min(Math.max(nextPage, 1), totalPages))
  }

  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent className={sideDrawerContentClassName('sm:max-w-5xl')}>
        <SheetHeader className={sideDrawerHeaderClassName()}>
          <SheetTitle>{t('Purchased Users')}</SheetTitle>
          <SheetDescription>{plan?.title || '-'}</SheetDescription>
        </SheetHeader>

        <div className={sideDrawerFormClassName('gap-4')}>
          <div className='flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
            <form className='flex min-w-0 flex-1 gap-2' onSubmit={submitSearch}>
              <Input
                value={keywordInput}
                onChange={(event) => setKeywordInput(event.target.value)}
                placeholder={t('Search users by name, email, or ID')}
                className='min-w-0 flex-1'
              />
              <Button type='submit' variant='outline'>
                <Search className='h-4 w-4' />
                <span className='sr-only'>{t('Search')}</span>
              </Button>
            </form>
            <div className='flex shrink-0 items-center gap-2'>
              <Select value={status} onValueChange={changeStatus}>
                <SelectTrigger className='w-36'>
                  <SelectValue placeholder={t('Status')} />
                </SelectTrigger>
                <SelectContent alignItemWithTrigger={false}>
                  <SelectGroup>
                    <SelectItem value='all'>{t('All statuses')}</SelectItem>
                    <SelectItem value='active'>{t('Active')}</SelectItem>
                    <SelectItem value='expired'>{t('Expired')}</SelectItem>
                    <SelectItem value='cancelled'>{t('Cancelled')}</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
              <Button
                type='button'
                variant='outline'
                onClick={() => query.refetch()}
                disabled={query.isFetching}
              >
                <RefreshCw
                  className={
                    query.isFetching ? 'h-4 w-4 animate-spin' : 'h-4 w-4'
                  }
                />
                <span className='sr-only'>{t('Refresh')}</span>
              </Button>
            </div>
          </div>

          <div className='min-h-0 flex-1 overflow-hidden rounded-md border'>
            <div className='hidden md:block'>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('User')}</TableHead>
                    <TableHead>{t('Status')}</TableHead>
                    <TableHead>{t('Validity')}</TableHead>
                    <TableHead>{t('Usage')}</TableHead>
                    <TableHead>{t('Source')}</TableHead>
                    <TableHead>{t('Created At')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {query.isLoading ? (
                    <TableRow>
                      <TableCell colSpan={6} className='py-10 text-center'>
                        {t('Loading...')}
                      </TableCell>
                    </TableRow>
                  ) : query.isError ? (
                    <TableRow>
                      <TableCell
                        colSpan={6}
                        className='text-destructive py-10 text-center'
                      >
                        {t('Loading failed')}
                      </TableCell>
                    </TableRow>
                  ) : records.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className='py-10'>
                        <Empty className='border-0'>
                          <EmptyHeader>
                            <EmptyTitle>{t('No purchase records')}</EmptyTitle>
                          </EmptyHeader>
                        </Empty>
                      </TableCell>
                    </TableRow>
                  ) : (
                    records.map((record) => {
                      const sub = record.subscription
                      return (
                        <TableRow key={sub.id}>
                          <TableCell className='max-w-72'>
                            <BuyerSummary record={record} />
                          </TableCell>
                          <TableCell>
                            <SubscriptionStatusBadge subscription={sub} />
                          </TableCell>
                          <TableCell>
                            <div className='text-xs'>
                              <div>{formatTimestamp(sub.start_time)}</div>
                              <div>{formatTimestamp(sub.end_time)}</div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <UsageText subscription={sub} />
                          </TableCell>
                          <TableCell>{sub.source || '-'}</TableCell>
                          <TableCell>
                            {formatTimestamp(record.user.created_at)}
                          </TableCell>
                        </TableRow>
                      )
                    })
                  )}
                </TableBody>
              </Table>
            </div>

            <div className='md:hidden'>
              {query.isLoading ? (
                <div className='py-10 text-center text-sm'>
                  {t('Loading...')}
                </div>
              ) : query.isError ? (
                <div className='text-destructive py-10 text-center text-sm'>
                  {t('Loading failed')}
                </div>
              ) : records.length === 0 ? (
                <Empty className='border-0 py-10'>
                  <EmptyHeader>
                    <EmptyTitle>{t('No purchase records')}</EmptyTitle>
                  </EmptyHeader>
                </Empty>
              ) : (
                <div className='px-3'>
                  {records.map((record) => (
                    <MobileSubscriptionCard
                      key={record.subscription.id}
                      record={record}
                    />
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className='flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
            <div className='text-muted-foreground text-sm'>
              {t('Page {{page}} of {{totalPages}}', {
                page,
                totalPages,
              })}
              <span className='ml-2'>
                {t('Total')}: {total}
              </span>
            </div>
            <Pagination className='mx-0 w-auto justify-start sm:justify-end'>
              <PaginationContent>
                <PaginationItem>
                  <PaginationPrevious
                    href='#'
                    text={t('Previous')}
                    aria-disabled={page <= 1}
                    onClick={(event) => {
                      event.preventDefault()
                      if (page > 1) goToPage(page - 1)
                    }}
                  />
                </PaginationItem>
                <PaginationItem>
                  <PaginationNext
                    href='#'
                    text={t('Next')}
                    aria-disabled={page >= totalPages}
                    onClick={(event) => {
                      event.preventDefault()
                      if (page < totalPages) goToPage(page + 1)
                    }}
                  />
                </PaginationItem>
              </PaginationContent>
            </Pagination>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}
