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
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Eye, MessageSquare } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Dialog } from '@/components/dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '@/components/ui/table'
import { TitledCard } from '@/components/ui/titled-card'
import { cn } from '@/lib/utils'
import { getMyFeedback, listMyFeedback } from '../api'
import type {
	PublicFeedback,
	PublicFeedbackStatus,
	PublicFeedbackType,
} from '../types'
import { formatComplianceTimestamp, normalizePagePayload } from '../utils'

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

function getFeedbackStatusLabel(
	t: (key: string) => string,
	value: PublicFeedbackStatus | string
) {
	const labels: Record<string, string> = {
		pending: t('Pending'),
		processing: t('Processing'),
		resolved: t('Resolved'),
		closed: t('Closed'),
		rejected: t('Rejected'),
	}
	return labels[value] ?? value
}

function getFeedbackStatusBadgeClass(status: PublicFeedbackStatus | string) {
	if (status === 'pending') return 'border-amber-200 bg-amber-50 text-amber-700'
	if (status === 'processing') return 'border-sky-200 bg-sky-50 text-sky-700'
	if (status === 'resolved')
		return 'border-emerald-200 bg-emerald-50 text-emerald-700'
	if (status === 'rejected') return 'border-rose-200 bg-rose-50 text-rose-700'
	return 'border-muted bg-muted text-muted-foreground'
}

function DetailField(props: { label: string; value?: string | number }) {
	return (
		<div className='border-b border-dashed py-2 last:border-b-0'>
			<div className='text-muted-foreground mb-1 text-xs'>{props.label}</div>
			<div className='break-all whitespace-pre-wrap text-sm'>
				{props.value || '-'}
			</div>
		</div>
	)
}

export function PublicFeedbackStatusCard() {
	const { t } = useTranslation()
	const [detailId, setDetailId] = useState<number | null>(null)

	const feedbackQuery = useQuery({
		queryKey: ['public-feedback', 'me'],
		queryFn: () => listMyFeedback({ p: 1, page_size: 20 }),
	})

	const detailQuery = useQuery({
		queryKey: ['public-feedback', 'me', detailId],
		queryFn: () => getMyFeedback(detailId ?? 0),
		enabled: Boolean(detailId),
	})

	const feedback = normalizePagePayload(feedbackQuery.data?.data).items
	const detailRecord = detailQuery.data?.data
	const closeDetail = () => setDetailId(null)

	return (
		<TitledCard
			title={t('My public feedback')}
			description={t('View complaint and feedback handling results.')}
			icon={<MessageSquare className='size-4' />}
		>
			<div className='space-y-2'>
				<Table>
					<TableHeader>
						<TableRow>
							<TableHead>{t('Tracking code')}</TableHead>
							<TableHead>{t('Type')}</TableHead>
							<TableHead>{t('Status')}</TableHead>
							<TableHead>{t('Title')}</TableHead>
							<TableHead>{t('Handled at')}</TableHead>
							<TableHead>{t('Processing note')}</TableHead>
							<TableHead className='text-right'>{t('Actions')}</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{feedbackQuery.isLoading ? (
							<TableRow>
								<TableCell colSpan={7}>
									<Skeleton className='h-10 w-full' />
								</TableCell>
							</TableRow>
						) : feedback.length === 0 ? (
							<TableRow>
								<TableCell
									colSpan={7}
									className='text-muted-foreground h-16 text-center'
								>
									{t('No feedback yet')}
								</TableCell>
							</TableRow>
						) : (
							feedback.map((record: PublicFeedback) => (
								<TableRow key={record.id}>
									<TableCell className='max-w-40 break-all'>
										{record.tracking_code}
									</TableCell>
									<TableCell>
										{getFeedbackTypeLabel(t, record.feedback_type)}
									</TableCell>
									<TableCell>
										<Badge
											variant='outline'
											className={cn(
												getFeedbackStatusBadgeClass(record.status)
											)}
										>
											{getFeedbackStatusLabel(t, record.status)}
										</Badge>
									</TableCell>
									<TableCell className='max-w-44 whitespace-normal'>
										{record.title}
									</TableCell>
									<TableCell>
										{formatComplianceTimestamp(record.handled_at)}
									</TableCell>
									<TableCell className='max-w-56 whitespace-normal'>
										{record.admin_note || '-'}
									</TableCell>
									<TableCell className='text-right'>
										<Button
											variant='outline'
											size='sm'
											onClick={() => setDetailId(record.id)}
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
			</div>

			<Dialog
				open={Boolean(detailId)}
				onOpenChange={(open) => {
					if (!open) closeDetail()
				}}
				title={t('Feedback details')}
				footer={
					<Button variant='outline' onClick={closeDetail}>
						{t('Close')}
					</Button>
				}
				contentClassName='sm:max-w-2xl'
			>
				{detailQuery.isLoading || !detailRecord ? (
					<Skeleton className='h-48 w-full' />
				) : (
					<div className='grid gap-3 sm:grid-cols-2'>
						<DetailField
							label={t('Tracking code')}
							value={detailRecord.tracking_code}
						/>
						<DetailField
							label={t('Type')}
							value={getFeedbackTypeLabel(t, detailRecord.feedback_type)}
						/>
						<DetailField
							label={t('Status')}
							value={getFeedbackStatusLabel(t, detailRecord.status)}
						/>
						<DetailField
							label={t('Submitted at')}
							value={formatComplianceTimestamp(detailRecord.created_at)}
						/>
						<DetailField
							label={t('Handled at')}
							value={formatComplianceTimestamp(detailRecord.handled_at)}
						/>
						<DetailField
							label={t('Title')}
							value={detailRecord.title}
						/>
						<div className='sm:col-span-2'>
							<DetailField label={t('Content')} value={detailRecord.content} />
						</div>
						<div className='sm:col-span-2'>
							<DetailField
								label={t('Processing note')}
								value={detailRecord.admin_note}
							/>
						</div>
					</div>
				)}
			</Dialog>
		</TitledCard>
	)
}
