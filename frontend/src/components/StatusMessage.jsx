export default function StatusMessage({ message, isError }) {
  return (
    <div className={`status ${isError ? 'error' : 'success'}`} role="status">
      {message}
    </div>
  )
}
