import { useToast } from 'vue-toastification'
import NotificationToast from '@/components/NotificationToast.vue'

class NotificationService {
  constructor() {
    this.toast = useToast()
  }

  // IP Blocked
  blocked(ip, score, reason, alertCount) {
    this.toast({
      component: NotificationToast,
      props: {
        type: 'block',
        title: 'IP Blocked',
        message: `${ip} has been automatically blocked`,
        ip,
        score,
        reason: reason || `${alertCount} suspicious alerts detected`
      }
    }, {
      timeout: 8000,
      icon: false
    })
  }

  // IP Unblocked
  unblocked(ip) {
    this.toast({
      component: NotificationToast,
      props: {
        type: 'unblock',
        title: 'IP Unblocked',
        message: `${ip} has been removed from blocklist`,
        ip
      }
    }, {
      timeout: 4000,
      icon: false
    })
  }

  // Generic success
  success(title, message) {
    this.toast.success({
      component: NotificationToast,
      props: {
        type: 'success',
        title,
        message
      }
    }, {
      timeout: 3000,
      icon: false
    })
  }

  // Generic error
  error(title, message) {
    this.toast.error({
      component: NotificationToast,
      props: {
        type: 'error',
        title,
        message
      }
    }, {
      timeout: 5000,
      icon: false
    })
  }

  // Generic info
  info(title, message) {
    this.toast.info({
      component: NotificationToast,
      props: {
        type: 'info',
        title,
        message
      }
    }, {
      timeout: 4000,
      icon: false
    })
  }
}

export default new NotificationService()
