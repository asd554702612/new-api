/*
Copyright (C) 2025 QuantumNous

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

import React, { useEffect, useState } from 'react';
import { Checkbox, Form, Modal } from '@douyinfe/semi-ui';

const PaymentRefundModal = ({ visible, order, mode = 'admin', loading, onCancel, onSubmit, t }) => {
  const [values, setValues] = useState({
    amount: 0,
    reason: '',
    force: false,
    deduct_balance: false,
  });

  useEffect(() => {
    if (visible && order) {
      setValues({
        amount: Number(order.pay_amount || 0),
        reason: '',
        force: false,
        deduct_balance: false,
      });
    }
  }, [visible, order]);

  const submit = () => {
    if (!values.reason.trim()) {
      return;
    }
    onSubmit(values);
  };

  return (
    <Modal
      title={mode === 'admin' ? t('登记退款') : t('申请退款')}
      visible={visible}
      onCancel={onCancel}
      onOk={submit}
      confirmLoading={loading}
      okText={mode === 'admin' ? t('登记') : t('提交')}
    >
      <Form layout='vertical'>
        {mode === 'admin' && (
          <Form.InputNumber
            label={t('退款金额')}
            field='amount'
            value={values.amount}
            min={0.01}
            precision={2}
            onChange={(amount) => setValues((prev) => ({ ...prev, amount }))}
          />
        )}
        <Form.TextArea
          label={t('退款原因')}
          field='reason'
          value={values.reason}
          autosize
          onChange={(reason) => setValues((prev) => ({ ...prev, reason }))}
        />
        {mode === 'admin' && (
          <div className='flex flex-col gap-2'>
            <Checkbox
              checked={values.deduct_balance}
              onChange={(event) =>
                setValues((prev) => ({
                  ...prev,
                  deduct_balance: event.target.checked,
                }))
              }
            >
              {t('同步扣减用户余额')}
            </Checkbox>
            <Checkbox
              checked={values.force}
              onChange={(event) =>
                setValues((prev) => ({ ...prev, force: event.target.checked }))
              }
            >
              {t('允许超过订单金额')}
            </Checkbox>
          </div>
        )}
      </Form>
    </Modal>
  );
};

export default PaymentRefundModal;
